package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	rendermodel "github.com/ONSdigital/dp-frontend-renderer/model"
	"github.com/ONSdigital/go-ns/log"
)

const errorPrefix string = "resolver."

var ZebedeeService zebedee.Service

// Resolve will take a URL and return a resolved version of the data.
func Resolve(uri string) ([]byte, error) {

	if uri == "/" {
		fmt.Println("Requested home_page...")

		var fullHomePage []byte
		var err error

		zebedeePageBytes, _, _ := ZebedeeService.GetData(uri)

		var homepage zebedeeModel.HomePage
		json.Unmarshal(zebedeePageBytes, &homepage)

		fullHomePage, err = json.Marshal(homepage)
		if err != nil {
			log.ErrorC("Failed to marshal full page.", err, nil)
			return fullHomePage, err
		}
		return fullHomePage, nil
	}

	zebedeeData, pageType, err := ZebedeeService.GetData(uri)
	if err != nil {
		return nil, err
	}

	if pageType == "home_page" {
		var page zebedeeModel.HomePage
		json.Unmarshal(zebedeeData, &page)

		// read the url from each section
		// create an array of the uris to resolve
		var urlsToResolve = make([]string, len(page.Sections))
		for i, section := range page.Sections {
			var section = zebedeeModel.HomeSection(section)
			log.Debug("Found uri to resolve in section", log.Data{"uri": section.Statistics.URI})
			urlsToResolve[i] = section.Statistics.URI
		}

		urlToChannelMap := make(map[string]chan zebedeeModel.HomePage)

		// get the json for each url (concurrently)
		for _, url := range urlsToResolve {
			channel := make(chan zebedeeModel.HomePage)
			urlToChannelMap[url] = channel
			go resolve(channel, url)
		}

		for _, channel := range urlToChannelMap {
			var page zebedeeModel.HomePage = <-channel
			fmt.Println("woot " + page.Description.Title)
		}

	} else {
		log.Debug("Page type not recognised: "+pageType, log.Data{})
	}

	return nil, nil
}

func resolve(ch chan zebedeeModel.HomePage, url string) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, pageType, err := ZebedeeService.GetData(url)
	if err != nil {
		log.Error(err, log.Data{})
		close(ch)
		return
	}

	var page zebedeeModel.HomePage
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

func resolveTaxonomy(uri string) ([]rendermodel.TaxonomyNode, error) {
	var rendererTaxonomyList []rendermodel.TaxonomyNode
	var zebedeeContentNodeList []zebedeeModel.ContentNode
	var zebedeeContentNodeBytes []byte

	zebedeeContentNodeBytes, err := ZebedeeService.GetTaxonomy(uri, 2)

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

func resolveParents(c chan []rendermodel.TaxonomyNode, uri string) {
	var err error
	zebedeeBytes, err := ZebedeeService.GetParents(uri)

	if err != nil {
		log.Error(fmtError("resolveParents", "Error getting parents from zebdee"), nil)
		close(c)
		return
	}

	var contentNodes []zebedeeModel.ContentNode
	err = json.Unmarshal(zebedeeBytes, &contentNodes)

	if err != nil {
		log.Error(fmtError("resolveParents", "Error unmarshalling zebedee parents json to contentNode."), nil)
		close(c)
		return
	}
	c <- contentNodeListToTaxonomyList(contentNodes)
	close(c)
}

func contentNodeListToTaxonomyList(contentNodes []zebedeeModel.ContentNode) (taxonomyList []rendermodel.TaxonomyNode) {
	for _, zebedeeContentNode := range contentNodes {
		taxonomyList = append(taxonomyList, zebedeeContentNode.Map())
	}
	return taxonomyList
}

func fmtError(funcName string, message string) error {
	return errors.New(fmt.Sprintf("%v%v: %v", errorPrefix, funcName, message))
}
