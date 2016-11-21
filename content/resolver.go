package content

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-content-resolver/babbage"
	"github.com/ONSdigital/dp-content-resolver/content/homePage"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/go-ns/common"
)

var pageTypeToResolver = map[string]func(*http.Request, zebedeeModel.HomePage, zebedee.Service, babbage.Service) ([]byte, error){
	"home_page": homePage.Resolve,
}

var ZebedeeService zebedee.Service
var BabbageService babbage.Service

// Resolve will take a URL and return a resolved version of the data.
func Resolve(req *http.Request) ([]byte, *common.ONSError) {
	uri := req.URL.Path

	zebedeeData, pageType, err := ZebedeeService.GetData(uri)
	if err != nil {
		return nil, err
	}

	// look up the resolver function from the pre-populated map.
	resolveFunc := pageTypeToResolver[pageType]

	if resolveFunc == nil {
		return nil, nil
	}

	var pageToResolve zebedeeModel.HomePage // zebedee model
	json.Unmarshal(zebedeeData, &pageToResolve)

	if pageToResolve.URI == "" {
		pageToResolve.URI = "/"
	}

	resolvedData, error := resolveFunc(req, pageToResolve, ZebedeeService, BabbageService)
	if err != nil {
		return nil, common.NewONSError(error, "Resolve error...")
	}
	return resolvedData, nil
}
