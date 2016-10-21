package resolvers

import (
	"encoding/json"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	"github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/go-ns/log"
)

// Resolver provides the interface for resolving data.
type Resolver interface {
	Resolve(url string) ([]byte, error)
}

// ZebedeeResolver is the zebedee specific implementation of the Resolver interface
type ZebedeeResolver struct {
	ZebedeeClient zebedee.Client
}

// Resolve will take a URL and return a resolved version of the data.
func (resolver ZebedeeResolver) Resolve(uri string) ([]byte, error) {

	if uri == "/" {
		log.Debug("Returning homepage stub data", nil)
		return []byte(stubbedData), nil
	}

	zebedeeData, pageType, err := resolver.ZebedeeClient.GetData(uri)
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

		//// get the json for each url (concurrently)
		//urlToChannelMap := make(map[string]chan model.HomePage)
		//
		//for _, url := range urlsToResolve {
		//	channel := make(chan model.HomePage)
		//	urlToChannelMap[url] = channel
		//	go resolve(channel, url)
		//}
		//
		////fmt.Println(page)
		//
		//for _, channel := range urlToChannelMap {
		//	var page unresolved.HomePage = <-channel
		//	fmt.Println("woot " + page.Description.Title)
		//}

	} else {
		log.Debug("Page type not recognised: "+pageType, log.Data{})
	}

	return nil, nil
}

func (resolver *ZebedeeResolver) resolve(ch chan model.HomePage, url string) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, pageType, err := resolver.ZebedeeClient.GetData(url)
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
