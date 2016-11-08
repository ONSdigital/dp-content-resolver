package content

import (
	"errors"
	"github.com/ONSdigital/dp-content-resolver/content/homePage"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
)

var pageTypeToResolver = map[string]func([]byte, zebedee.Service) ([]byte, error){
	"home_page": homePage.Resolve,
}

var ZebedeeService zebedee.Service

// Resolve will take a URL and return a resolved version of the data.
func Resolve(uri string) ([]byte, error) {

	zebedeeData, pageType, err := ZebedeeService.GetData(uri)
	if err != nil {
		return nil, err
	}

	// look up the resolver function from the pre-populated map.
	resolveFunc := pageTypeToResolver[pageType]

	if resolveFunc == nil {
		return nil, errors.New("Page type not recognised: " + pageType)
	}

	resolvedData, err := resolveFunc(zebedeeData, ZebedeeService)
	if err != nil {
		return nil, err
	}
	return resolvedData, err
}
