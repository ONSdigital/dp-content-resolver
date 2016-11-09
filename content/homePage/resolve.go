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
const taxonomyErrMsg string = "Unexpected error while resolving page taxonomy. Will attempt to render without taxonomy."
const breadcrumbErrMsg string = "Unexpected error while resolving page breadcrumbs. Will attempt to render without breadcrumbs."
const headlineErrMsg string = "Error while trying to resolve page section: URI: %s"

// Resolve the given page data.
func Resolve(targetUri string, pageToResolve zebedeeModel.HomePage, zebedeeService zebedee.Service) (resolvedPageData []byte, err error) {
    var resolvedPage homepage.Page = homepage.Page{URI: targetUri}
    var taxonomyErr error
    var breadcrumbErr error
    var headlinesErr = make(map[string]error, 0)

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
        resolvedPage.Data.HeadlineFigures, headlinesErr = resolveSections(pageToResolve.Sections, zebedeeService)
        wg.Done()
    }()

    wg.Wait() // wait for all the resolve jobs to complete.

    if taxonomyErr != nil {
        log.ErrorC(taxonomyErrMsg, taxonomyErr, nil)
        resolvedPage.Taxonomy = make([]renderModel.TaxonomyNode, 0)
    }

    if breadcrumbErr != nil {
        log.ErrorC(breadcrumbErrMsg, breadcrumbErr, nil)
        resolvedPage.Breadcrumb = make([]renderModel.TaxonomyNode, 0)
    }

    for uri, err := range headlinesErr {
        message := fmt.Sprintf(headlineErrMsg, uri)
        log.ErrorC(message, err, nil)
    }

    resolvedPageData, err = json.Marshal(resolvedPage)
    return
}

func resolveSections(pageSections []*zebedeeModel.HomeSection, zebedeeService zebedee.Service) ([]*homepage.HeadlineFigure, map[string]error) {
    headlines := make([]*homepage.HeadlineFigure, len(pageSections))
    headlineErrors := make(map[string]error, 0)

    wg := new(sync.WaitGroup)
    wg.Add(len(pageSections))

    for i, section := range pageSections {
        go func(index int, section *zebedeeModel.HomeSection) {
            var timeSeriesPage *zebedeeModel.TimeseriesPage
            var err error
            timeSeriesPage, err = getTimeSeriesPage(section.Statistics.URI, zebedeeService)

            if err != nil {
                log.Error(err, log.Data{"message": "failed to get timeseries page for URL: " + section.Statistics.URI})
                headlineErrors[section.Statistics.URI] = err
            } else {
                // assigning via index rather than append() means we wont get concurrency issues.
                headlines[index] = mapTimeseriesToHeadlineFigure(timeSeriesPage)
            }
            wg.Done()
        }(i, section)
    }
    wg.Wait()

    // If there were any errors there will be nil values in the results remove these before returning.
    var results []*homepage.HeadlineFigure
    for _, h := range headlines {
        if h != nil {
            results = append(results, h)
        }
    }
    return results, headlineErrors
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

// Resolve the breadcrumbs for this page.
func resolveParents(uri string, zebedeeService zebedee.Service) ([]renderModel.TaxonomyNode, error) {
    var err error
    var taxonomyList []renderModel.TaxonomyNode
    fmt.Println("\ngetting Parents\n")
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
    return taxonomyList, nil
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
