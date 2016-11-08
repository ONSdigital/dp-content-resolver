package homePage

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/dp-frontend-renderer/handlers/homepage"
	"github.com/ONSdigital/go-ns/log"
	"sync"
)

// Resolve the given page data.
func Resolve(zebedeeData []byte, getData func(url string) (data []byte, pageType string, err error)) (resolvedPageData []byte, err error) {
	var pageToResolve model.HomePage // zebedee model
	var resolvedPage homepage.Page

	// Todo - remove this when all the things are resolved. Using the stubbed data as a basis while not all
	// parts of the page are being resolved.
	json.Unmarshal([]byte(stubbedData), &resolvedPage)
	json.Unmarshal(zebedeeData, &pageToResolve)

	var wg sync.WaitGroup // synchronises each resolve job running concurrently.

	resolveSections(&pageToResolve, &resolvedPage, wg, getData)

	wg.Wait() // wait for all the resolve jobs to complete.

	resolvedPageData, err = json.Marshal(resolvedPage)
	return
}

func resolveSections(pageToResolve *model.HomePage, resolvedPage *homepage.Page, wg sync.WaitGroup,
	getData func(url string) (data []byte, pageType string, err error)) {

	// initialise the headline figures array allowing each index to be assigned to concurrently.
	resolvedPage.Data.HeadlineFigures = make([]*homepage.HeadlineFigure, len(pageToResolve.Sections))

	for i, section := range pageToResolve.Sections {
		wg.Add(1)

		go func(i int, section *model.HomeSection) {
			defer wg.Done()

			url := section.Statistics.URI
			timeSeriesPage, err := getTimeSeriesPage(url, getData)
			if err != nil {
				log.Error(err, log.Data{"message": "failed to get timeseries page for URL: " + url})
				return
			}

			fmt.Println(resolvedPage.Data.HeadlineFigures[i])
			resolvedPage.Data.HeadlineFigures[i] = mapTimeseriesToHeadlineFigure(timeSeriesPage)
			fmt.Println(resolvedPage.Data.HeadlineFigures[i])

		}(i, section)
	}
}

func getTimeSeriesPage(url string, getData func(url string) (data []byte, pageType string, err error)) (*model.TimeseriesPage, error) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, _, err := getData(url + "&series")
	if err != nil {
		log.Error(err, log.Data{})
		return nil, err
	}

	var page *model.TimeseriesPage
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Error(err, nil)
		return nil, err
	}

	return page, nil
}

func mapTimeseriesToHeadlineFigure(page *model.TimeseriesPage) (figure *homepage.HeadlineFigure) {

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
