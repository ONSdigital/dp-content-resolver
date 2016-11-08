package homePage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/dp-frontend-renderer/handlers/homepage"
	renderModel "github.com/ONSdigital/dp-frontend-renderer/model"
	"github.com/ONSdigital/go-ns/log"
	"sync"
)

const errorPrefix string = "homepage.resolver."

// Resolve the given page data.
func Resolve(zebedeeData []byte, zebedeeService zebedee.Service) (resolvedPageData []byte, err error) {
	var pageToResolve zebedeeModel.HomePage // zebedee model
	var resolvedPage homepage.Page

	// Todo - remove this when all the things are resolved. Using the stubbed data as a basis while not all
	// parts of the page are being resolved.
	json.Unmarshal([]byte(stubbedData), &resolvedPage)
	json.Unmarshal(zebedeeData, &pageToResolve)

	var wg sync.WaitGroup // synchronises each resolve job running concurrently.

	resolveTaxonomyAsync(pageToResolve.URI, &resolvedPage, wg, zebedeeService)
	//resolveParentsAsync(pageToResolve.URI, &resolvedPage, wg, zebedeeService) // breadcrumb
	//resolveSections(&pageToResolve, &resolvedPage, wg, zebedeeService)

	wg.Wait() // wait for all the resolve jobs to complete.

	resolvedPageData, err = json.Marshal(resolvedPage)
	return
}

func resolveTaxonomyAsync(uri string, resolvedPage *homepage.Page, wg sync.WaitGroup, zebedeeService zebedee.Service) {
	wg.Add(1)

	go func() {
		var taxonomy []renderModel.TaxonomyNode
		taxonomy, _ = resolveTaxonomy(uri, zebedeeService)
		resolvedPage.Taxonomy = &taxonomy

		fmt.Printf("%+v\n", resolvedPage)
		wg.Done()
	}()
}

func resolveParentsAsync(uri string, resolvedPage *homepage.Page, wg sync.WaitGroup, zebedeeService zebedee.Service) {
	wg.Add(1)

	go func() {
		var breadcrumb []renderModel.TaxonomyNode
		breadcrumb, _ = resolveParents(uri, zebedeeService)
		resolvedPage.Breadcrumb = breadcrumb
		wg.Done()
	}()
}

func resolveSections(pageToResolve *zebedeeModel.HomePage, resolvedPage *homepage.Page, wg sync.WaitGroup, zebedeeService zebedee.Service) {

	// initialise the headline figures array allowing each index to be assigned to concurrently.
	resolvedPage.Data.HeadlineFigures = make([]*homepage.HeadlineFigure, len(pageToResolve.Sections))

	for i, section := range pageToResolve.Sections {
		wg.Add(1)

		go func(i int, section *zebedeeModel.HomeSection, resolvedPage *homepage.Page) {
			defer wg.Done()

			url := section.Statistics.URI
			timeSeriesPage, err := getTimeSeriesPage(url, zebedeeService)
			if err != nil {
				log.Error(err, log.Data{"message": "failed to get timeseries page for URL: " + url})
				return
			}

			//fmt.Println(resolvedPage.Data.HeadlineFigures[i])
			resolvedPage.Data.HeadlineFigures[i] = mapTimeseriesToHeadlineFigure(timeSeriesPage)
			//fmt.Println(resolvedPage.Data.HeadlineFigures[i])

		}(i, section, resolvedPage)
	}
}

func getTimeSeriesPage(url string, zebedeeService zebedee.Service) (*zebedeeModel.TimeseriesPage, error) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, _, err := zebedeeService.GetData(url + "&series")
	if err != nil {
		log.Error(err, log.Data{})
		return nil, err
	}

	var page *zebedeeModel.TimeseriesPage
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Error(err, nil)
		return nil, err
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

func resolveTaxonomy(uri string, zebedeeService zebedee.Service) ([]renderModel.TaxonomyNode, error) {
	var rendererTaxonomyList []renderModel.TaxonomyNode
	var zebedeeContentNodeList []zebedeeModel.ContentNode
	var zebedeeContentNodeBytes []byte

	zebedeeContentNodeBytes, err := zebedeeService.GetTaxonomy(uri, 2)

	if err != nil {
		log.ErrorC("resolver.resolveTaxonomy: Error resolving taxonomy", err, nil)
		return rendererTaxonomyList, err
	}

	err = json.Unmarshal(zebedeeContentNodeBytes, &zebedeeContentNodeList)
	if err != nil {
		log.ErrorC("resolver.resolveTaxonomy: Error unmarshalling json to taxonomy node.", err, nil)
		return rendererTaxonomyList, fmtError("resolveTaxonomy", "Error unmarshalling taxonomy json.")
	}
	return contentNodeListToTaxonomyList(zebedeeContentNodeList), nil
}

func resolveParents(uri string, zebedeeService zebedee.Service) ([]renderModel.TaxonomyNode, error) {
	var err error
	var taxonomyList []renderModel.TaxonomyNode
	zebedeeBytes, err := zebedeeService.GetParents(uri)

	if err != nil {
		log.Error(fmtError("resolveParents", "Error getting parents from zebdee"), nil)
		return taxonomyList, err
	}

	var contentNodes []zebedeeModel.ContentNode
	err = json.Unmarshal(zebedeeBytes, &contentNodes)

	if err != nil {
		log.Error(fmtError("resolveParents", "Error unmarshalling zebedee parents json to contentNode."), nil)
		return taxonomyList, err
	}

	taxonomyList = contentNodeListToTaxonomyList(contentNodes)
	return taxonomyList, err
}

func contentNodeListToTaxonomyList(contentNodes []zebedeeModel.ContentNode) (taxonomyList []renderModel.TaxonomyNode) {
	for _, zebedeeContentNode := range contentNodes {
		taxonomyList = append(taxonomyList, zebedeeContentNode.Map())
	}
	return taxonomyList
}

func fmtError(funcName string, message string) error {
	return errors.New(fmt.Sprintf("%v%v: %v", errorPrefix, funcName, message))
}
