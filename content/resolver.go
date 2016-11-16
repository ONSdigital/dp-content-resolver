package content

import (
    "errors"
    "github.com/ONSdigital/dp-content-resolver/content/homePage"
    "github.com/ONSdigital/dp-content-resolver/zebedee"
    "encoding/json"
    zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
)

var pageTypeToResolver = map[string]func(string, zebedeeModel.HomePage, zebedee.Service) ([]byte, error){
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

    var pageToResolve zebedeeModel.HomePage // zebedee model
    json.Unmarshal(zebedeeData, &pageToResolve)

    resolvedData, err := resolveFunc(uri, pageToResolve, ZebedeeService)
    if err != nil {
        return nil, err
    }
    return resolvedData, err
}
