package homePage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ONSdigital/dp-content-resolver/zebedee"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	renderModel "github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/homepage"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

type ResolvedHeadlines []*ResolvedHeadline

type ResolvedHeadline struct {
	headline *homepage.HeadlineFigure
	err      error
	meta     log.Data
}

func (r ResolvedHeadlines) countErrors() int {
	count := 0
	for _, i := range r {
		if i.isError() {
			count++
		}
	}
	return count
}

func (r ResolvedHeadlines) getFailures() []*ResolvedHeadline {
	failures := make([]*ResolvedHeadline, 0)
	for _, item := range r {
		if item.isError() {
			failures = append(failures, item)
		}
	}
	return failures
}

func (r *ResolvedHeadline) isError() bool {
	return r.err != nil
}

// Resolve the given page data.
func Resolve(req *http.Request, pageToResolve zebedeeModel.HomePage, zebedeeService zebedee.Service) (resolvedPageData []byte, err error) {
	var resolvedPage homepage.Page = homepage.Page{URI: pageToResolve.URI}
	var taxonomyErr *common.ONSError
	var breadcrumbErr *common.ONSError
	var headlines ResolvedHeadlines

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go func() {
		resolvedPage.Taxonomy, taxonomyErr = resolveTaxonomy(resolvedPage.URI, zebedeeService)
		wg.Done()
	}()

	go func() {
		resolvedPage.Breadcrumb, breadcrumbErr = resolveParents(resolvedPage.URI, zebedeeService)
		wg.Done()
	}()

	go func() {
		headlines = resolveHeadlineSections(pageToResolve.Sections, zebedeeService)
		wg.Done()
	}()

	wg.Wait() // wait for all the resolve jobs to complete.

	if taxonomyErr != nil {
		log.ErrorR(req, taxonomyErr, nil)
	}

	if breadcrumbErr != nil {
		log.ErrorR(req, breadcrumbErr, nil)
	}

	if errorCount := headlines.countErrors(); errorCount > 0 {
		log.ErrorR(req, fmt.Errorf("One of more headline sections failed to resolve."), log.Data{
			"totalHeadLineResolves":  len(pageToResolve.Sections) + 1,
			"failedHeadLineResolves": errorCount,
		})
	}

	resolvedPage.Data.HeadlineFigures = make([]*homepage.HeadlineFigure, 0)
	for _, resolvedItem := range headlines {
		if resolvedItem.isError() {
			log.ErrorR(req, resolvedItem.err, resolvedItem.meta)
		} else {
			resolvedPage.Data.HeadlineFigures = append(resolvedPage.Data.HeadlineFigures, resolvedItem.headline)
		}
	}
	resolvedPageData, err = json.Marshal(resolvedPage)
	return
}

func resolveHeadlineSections(pageSections []*zebedeeModel.HomeSection, zebedeeService zebedee.Service) ResolvedHeadlines {
	results := make(ResolvedHeadlines, len(pageSections))
	wg := new(sync.WaitGroup)
	wg.Add(len(pageSections))

	for i, section := range pageSections {
		go func(index int, section *zebedeeModel.HomeSection) {
			var timeSeriesPage *zebedeeModel.TimeseriesPage
			var onsError *common.ONSError
			var result *ResolvedHeadline
			timeSeriesPage, onsError = getTimeSeriesPage(section.Statistics.URI, zebedeeService)

			if onsError != nil {
				onsError.AddParameter("resolveURI", section.Statistics.URI)
				onsError.AddParameter("description", "Failed to resolve headline section.")

				result = &ResolvedHeadline{
					err: onsError,
				}
			} else {
				result = &ResolvedHeadline{headline: mapTimeseriesToHeadlineFigure(timeSeriesPage)}
			}

			results[index] = result
			wg.Done()
		}(i, section)
	}
	wg.Wait()
	return results
}

func getTimeSeriesPage(uri string, zebedeeService zebedee.Service) (*zebedeeModel.TimeseriesPage, *common.ONSError) {
	data, err := zebedeeService.GetTimeSeries(uri)
	if err != nil {
		return nil, err
	}

	var page *zebedeeModel.TimeseriesPage
	unmarshalErr := json.Unmarshal(data, &page)
	if unmarshalErr != nil {
		return nil, common.NewONSError(unmarshalErr, "Error unmarshalling timeseries pages json.")
	}

	return page, nil
}

func mapTimeseriesToHeadlineFigure(page *zebedeeModel.TimeseriesPage) (figure *homepage.HeadlineFigure) {

	figure = &homepage.HeadlineFigure{
		Title: page.Description.Title,
	}

	figure.URI = page.URI
	figure.ReleaseDate = page.Description.ReleaseDate

	figure.LatestFigure = homepage.LatestFigure{
		PreUnit: page.Description.PreUnit,
		Unit:    page.Description.Unit,
		Figure:  page.Description.Number,
	}

	figure.SparklineData = make([]homepage.SparklineData, len(page.Series))
	for i, seriesItem := range page.Series {
		figure.SparklineData[i] = homepage.SparklineData{
			Name:    seriesItem.Name,
			StringY: seriesItem.StringY,
			Y:       seriesItem.Y,
		}
	}

	return figure
}

func resolveTaxonomy(uri string, zebedeeService zebedee.Service) ([]renderModel.TaxonomyNode, *common.ONSError) {
	var rendererTaxonomyList []renderModel.TaxonomyNode
	var zebedeeContentNodeList []zebedeeModel.ContentNode
	var zebedeeContentNodeBytes []byte

	zebedeeContentNodeBytes, err := zebedeeService.GetTaxonomy(uri, 2)

	if err != nil {
		return rendererTaxonomyList, err
	}

	unmarshallErr := json.Unmarshal(zebedeeContentNodeBytes, &zebedeeContentNodeList)
	if unmarshallErr != nil {
		return rendererTaxonomyList, common.NewONSError(unmarshallErr, "Error while attempting to unmarshal content taxonomy nodes.")
	}
	return contentNodeListToTaxonomyList(zebedeeContentNodeList), nil
}

// Resolve the breadcrumbs for this page.
func resolveParents(uri string, zebedeeService zebedee.Service) ([]renderModel.TaxonomyNode, *common.ONSError) {
	var taxonomyList []renderModel.TaxonomyNode
	zebedeeBytes, err := zebedeeService.GetParents(uri)

	if err != nil {
		return taxonomyList, err
	}

	var contentNodes []zebedeeModel.ContentNode
	unmarshallErr := json.Unmarshal(zebedeeBytes, &contentNodes)

	if unmarshallErr != nil {
		return taxonomyList, common.NewONSError(unmarshallErr, "Error while attempting to unmarshal parent content nodes.")
	}

	taxonomyList = contentNodeListToTaxonomyList(contentNodes)
	return taxonomyList, nil
}

func contentNodeListToTaxonomyList(contentNodes []zebedeeModel.ContentNode) (taxonomyList []renderModel.TaxonomyNode) {
	for _, zebedeeContentNode := range contentNodes {
		taxonomyList = append(taxonomyList, zebedeeContentNode.Map())
	}
	return taxonomyList
}
