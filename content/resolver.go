package content

import (
	"errors"
	"github.com/ONSdigital/dp-content-resolver/content/homePage"
)

var pageTypeToResolver = map[string]func([]byte, func(url string) (data []byte, pageType string, err error)) ([]byte, error){
	"home_page": homePage.Resolve,
}

// GetData is a generic function definition allowing different
// implementations to be injected.
var GetData func(url string) (data []byte, pageType string, err error)

// Resolve will take a URL and return a resolved version of the data.
func Resolve(uri string) ([]byte, error) {

	zebedeeData, pageType, err := GetData(uri)
	if err != nil {
		return nil, err
	}

	// look up the resolver function from the pre-populated map.
	resolveFunc := pageTypeToResolver[pageType]

	if resolveFunc == nil {
		return nil, errors.New("Page type not recognised: " + pageType)
	}

	resolvedData, err := resolveFunc(zebedeeData, GetData)
	if err != nil {
		return nil, err
	}

	return resolvedData, nil
}
