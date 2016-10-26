package content

import (
	"encoding/json"
	"fmt"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	rendermodel "github.com/ONSdigital/dp-frontend-renderer/model"
	"github.com/ONSdigital/go-ns/log"
)

// GetData is a generic function definition allowing different
// implementations to be injected.
var GetData func(url string) (data []byte, pageType string, err error)

var GetTaxonomy func(url string, depth int) ([]byte, error)

// Resolve will take a URL and return a resolved version of the data.
func Resolve(uri string) ([]byte, error) {

	if uri == "/" {
		fmt.Println("Requested home_page...")

		var fullHomePage []byte
		var err error

		zebedeePageBytes, _, _ := GetData(uri)
		//rendererTaxonomyNodes, _ := resolveTaxonomy(uri)

		var homepage zebedeeModel.HomePage
		json.Unmarshal(zebedeePageBytes, &homepage)

		fullHomePage, err = json.Marshal(homepage)
		if err != nil {
			log.ErrorC("Failed to marshal full page.", err, nil)
			return fullHomePage, err
		}
		return fullHomePage, nil
	}

	zebedeeData, pageType, err := GetData(uri)
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
		log.Debug("Page type not recognised: " + pageType, log.Data{})
	}

	return nil, nil
}

func resolve(ch chan zebedeeModel.HomePage, url string) {
	log.Debug("Resolving page data", log.Data{"url": url})

	data, pageType, err := GetData(url)
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

func resolveTaxonomy(uri string) (rendererTaxonomyList []rendermodel.TaxonomyNode, err error) {
	rendererTaxonomyList = make([]rendermodel.TaxonomyNode, 0)
	var zebedeeTaxonomyList = make([]zebedeeModel.Taxonomy, 0)
	var zebedeeTaxonomyBytes []byte

	zebedeeTaxonomyBytes, err = GetTaxonomy(uri, 2)

	if err != nil {
		log.ErrorC("resolver.resolveTaxonomy: Error resolving taxonomy", err, nil)
		return rendererTaxonomyList, err
	}

	err = json.Unmarshal(zebedeeTaxonomyBytes, &zebedeeTaxonomyList)
	if err != nil {
		log.ErrorC("resolver.resolveTaxonomy: Error unmarshalling json to taxonomy node.", err, nil)
		return rendererTaxonomyList, err
	}

	// Convert the from the zebedee model to the renderer model.
	rendererTaxonomyList = make([]rendermodel.TaxonomyNode, len(zebedeeTaxonomyList))
	for _, zebedeeTaxonomyNode := range zebedeeTaxonomyList {
		rendererTaxonomyList = append(rendererTaxonomyList, zebedeeTaxonomyNode.Map())
	}

	return rendererTaxonomyList, nil
}
