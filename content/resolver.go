package content

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/go-ns/log"
)

// GetData is a generic function definition allowing different
// implementations to be injected.
var GetData func(url string) (data []byte, pageType string, err error)

// Resolve will take a URL and return a resolved version of the data.
func Resolve(uri string) ([]byte, error) {

	if uri == "/" {
		// TODO REMOVE BEFORE COMMIT
		log.Debug("Using Stubbed JSON", nil)
		return []byte(stubbedData), nil
	}

	zebedeeData, pageType, err := GetData(uri)
	if err != nil {
		return nil, err
	}

	if pageType == "home_page" {
		var page model.HomePage
		json.Unmarshal(zebedeeData, &page)

		// read the url from each section
		// create an array of the uris to resolve
		var urlsToResolve = make([]string, len(page.Sections))
		for i, section := range page.Sections {
			var section = model.HomeSection(section)
			log.Debug("Found uri to resolve in section", log.Data{"uri": section.Statistics.URI})
			urlsToResolve[i] = section.Statistics.URI
		}

		urlToChannelMap := make(map[string]chan model.HomePage)

		// get the json for each url (concurrently)
		for _, url := range urlsToResolve {
			channel := make(chan model.HomePage)
			urlToChannelMap[url] = channel
			go resolve(channel, url)
		}

		for _, channel := range urlToChannelMap {
			var page model.HomePage = <-channel
			fmt.Println("woot " + page.Description.Title)
		}

	} else {
		log.Debug("Page type not recognised: " + pageType, log.Data{})
	}

	return nil, nil
}

func resolve(ch chan model.HomePage, url string) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, pageType, err := GetData(url)
	if err != nil {
		log.Error(err, log.Data{})
		close(ch)
		return
	}

	var page model.HomePage
	jsonErr := json.Unmarshal(data, &page)
	if jsonErr != nil {
		log.Error(jsonErr, nil)
		close(ch)
		return
	}

	log.Debug("Resolving page success", log.Data{"url": url, "pageType": pageType})

	ch <- page
	close(ch)
}
