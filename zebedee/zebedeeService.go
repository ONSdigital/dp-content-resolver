package zebedee

import (
	"github.com/ONSdigital/go-ns/common"
)

// Service defines interface of zebedee service.
type Service interface {
	GetData(url string, requestContentID string) (data []byte, pageType string, err *common.ONSError)
	GetTaxonomy(url string, depth int, requestContentID string) ([]byte, *common.ONSError)
	GetParents(url string, requestContentID string) ([]byte, *common.ONSError)
	GetTimeSeries(url string, requestContentID string) ([]byte, *common.ONSError)
}
