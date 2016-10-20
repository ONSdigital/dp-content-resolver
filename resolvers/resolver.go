package resolvers

import (
	"encoding/json"
	"fmt"
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

		// get the json for each url (concurrently)
		urlToChannelMap := make(map[string]chan model.HomePage)

		for _, url := range urlsToResolve {
			channel := make(chan model.HomePage)
			urlToChannelMap[url] = channel
			go resolve(channel, url)
		}

		//fmt.Println(page)

		for _, channel := range urlToChannelMap {
			var page unresolved.HomePage = <-channel
			fmt.Println("woot " + page.Description.Title)
		}

	} else {
		log.Debug("Page type not recognised: "+pageType, log.Data{})
	}

	return nil, nil
}

//func ResolveDoNotLook(url string) error {
//
//	response, error := zebedee.GetData(url);
//	if error != nil {
//		log.Error(error, log.Data{})
//	}
//
//	responseBody, err := ioutil.ReadAll(response.Body)
//	response.Body.Close()
//	if err != nil {
//		log.Error(err, nil)
//	}
//	//fmt.Println(response)
//	//fmt.Printf("%s", responseBody)
//
//	// unmarshal the page type field only.
//	var pageType unresolved.PageType
//	jsonErr := json.Unmarshal(responseBody, &pageType)
//	if jsonErr != nil {
//		log.Error(err, nil)
//	}
//	fmt.Println("Page type = " + pageType.Type);
//
//	var page unresolved.HomePage;
//	json.Unmarshal(responseBody, &page)
//
//	// create an array of the uris to resolve
//	var urlsToResolve = make([]string, len(page.Sections))
//	for i, section := range page.Sections {
//		var s = unresolved.HomeSection(section)
//		fmt.Println(s.Statistics.URI)
//		urlsToResolve[i] = s.Statistics.URI
//	}
//
//	urlToChannelMap := make(map[string]chan unresolved.HomePage)
//
//	for _, url := range urlsToResolve {
//		channel := make(chan unresolved.HomePage)
//		urlToChannelMap[url] = channel
//		go resolve(channel, url)
//	}
//
//	//fmt.Println(page)
//
//	for _, channel := range urlToChannelMap {
//		var page unresolved.HomePage = <-channel
//		fmt.Println("woot " + page.Description.Title)
//	}
//}
//
//
//func resolvewhatever(ch chan unresolved.HomePage, url string) {
//	fmt.Println("Resolving..." + url)
//
//	response, error := netClient.Get("http://localhost:8082/data?uri=" + url)
//	if error != nil {
//		log.Error(error, log.Data{})
//		close(ch)
//		return
//	}
//
//	responseBody, err := ioutil.ReadAll(response.Body)
//	response.Body.Close()
//	if err != nil {
//		log.Error(err, nil)
//		close(ch)
//		return
//	}
//
//	var page unresolved.HomePage;
//	jsonErr := json.Unmarshal(responseBody, &page)
//	if jsonErr != nil {
//		log.Error(jsonErr, nil)
//		close(ch)
//		return
//	}
//
//	fmt.Println("Resolve Done!" + page.Type)
//	ch <- page
//	close(ch)
//}
