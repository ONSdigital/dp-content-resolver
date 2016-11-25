package zebedee

import (
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"github.com/ONSdigital/go-ns/common"
)

// Service defines interface of zebedee service.
type Service interface {
	GetData(url string, requestContentID string) (data []byte, pageType string, err *common.ONSError)
	GetTaxonomy(url string, depth int, requestContentID string) ([]zebedeeModel.ContentNode, *common.ONSError)
	GetParents(url string, requestContentID string) ([]zebedeeModel.ContentNode, *common.ONSError)
	GetTimeSeries(url string, requestContentID string) (*zebedeeModel.TimeseriesPage, *common.ONSError)
}
