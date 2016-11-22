package homePage

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-content-resolver/requests"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	renderModel "github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/homepage"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"net/http"
	"sync"
)

var ZebedeeService zebedee.Service

const taxonomyLandingPageType = "taxonomy_landing_page"

type resolvedHeadlines []*resolvedHeadline

type resolvedHeadline struct {
	headline *homepage.HeadlineFigure
	err      error
	meta     log.Data
}

func (r resolvedHeadlines) countErrors() int {
	count := 0
	for _, i := range r {
		if i.isError() {
			count++
		}
	}
	return count
}

func (r resolvedHeadlines) getFailures() []*resolvedHeadline {
	failures := make([]*resolvedHeadline, 0)
	for _, item := range r {
		if item.isError() {
			failures = append(failures, item)
		}
	}
	return failures
}

func (r *resolvedHeadline) isError() bool {
	return r.err != nil
}

// Resolve the given page data.
func Resolve(req *http.Request, pageToResolve zebedeeModel.HomePage, reqContentIDGen requests.ContextIDGenerator) (resolvedPageData []byte, err error) {
	var resolvedPage = homepage.Page{URI: pageToResolve.URI}
	var taxonomyErr *common.ONSError
	var breadcrumbErr *common.ONSError
	var headlines resolvedHeadlines

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go func() {
		resolvedPage.Taxonomy, taxonomyErr = resolveTaxonomy(resolvedPage.URI, reqContentIDGen)
		wg.Done()
	}()

	go func() {
		resolvedPage.Breadcrumb, breadcrumbErr = resolveParents(resolvedPage.URI, reqContentIDGen)
		wg.Done()
	}()

	go func() {
		headlines = resolveHeadlineSections(pageToResolve.Sections, reqContentIDGen)
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

func resolveHeadlineSections(pageSections []*zebedeeModel.HomeSection, reqContextIDGen requests.ContextIDGenerator) resolvedHeadlines {
	results := make(resolvedHeadlines, len(pageSections))
	wg := new(sync.WaitGroup)
	wg.Add(len(pageSections))

	for i, section := range pageSections {
		go func(index int, section *zebedeeModel.HomeSection) {

			var timeSeriesPage *zebedeeModel.TimeseriesPage
			var onsError *common.ONSError
			var result *resolvedHeadline
			timeSeriesPage, onsError = ZebedeeService.GetTimeSeries(section.Statistics.URI, reqContextIDGen.Generate())

			if onsError != nil {
				onsError.AddParameter("resolveURI", section.Statistics.URI)
				onsError.AddParameter("description", "Failed to resolve headline section.")

				result = &resolvedHeadline{
					err:  onsError.RootError,
					meta: onsError.Parameters,
				}
			} else {
				result = &resolvedHeadline{headline: mapTimeseriesToHeadlineFigure(timeSeriesPage)}
			}

			results[index] = result
			wg.Done()
		}(i, section)
	}
	wg.Wait()
	return results
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

func resolveTaxonomy(uri string, reqContextIDGen requests.ContextIDGenerator) ([]renderModel.TaxonomyNode, *common.ONSError) {
	var rendererTaxonomyList []renderModel.TaxonomyNode
	zebedeeContentNodeList, err := ZebedeeService.GetTaxonomy(uri, 2, reqContextIDGen.Generate())

	if err != nil {
		return rendererTaxonomyList, err
	}

	for _, zebedeeContentNode := range zebedeeContentNodeList {
		if zebedeeContentNode.PageType == taxonomyLandingPageType {
			rendererTaxonomyList = append(rendererTaxonomyList, zebedeeContentNode.Map())
		}
	}
	return rendererTaxonomyList, nil
}

// resolveParents get the parents data from zebedee and convert it into the renderer model.
func resolveParents(uri string, reqContextIDGen requests.ContextIDGenerator) ([]renderModel.TaxonomyNode, *common.ONSError) {
	var taxonomyNodeList []renderModel.TaxonomyNode
	zebedeeContentNodes, err := ZebedeeService.GetParents(uri, reqContextIDGen.Generate())

	if err != nil {
		return taxonomyNodeList, err
	}

	for _, zebedeeContentNode := range zebedeeContentNodes {
		taxonomyNodeList = append(taxonomyNodeList, zebedeeContentNode.Map())
	}
	return taxonomyNodeList, nil
}
