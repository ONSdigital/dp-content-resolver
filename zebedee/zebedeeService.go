package zebedee

import (
    "github.com/ONSdigital/go-ns/common"
)

// Defines interface of zebedee service.
type Service interface {
    GetData(url string) (data []byte, pageType string, err *common.ONSError)
    GetTaxonomy(url string, depth int) ([]byte, *common.ONSError)
    GetParents(url string) ([]byte, *common.ONSError)
    GetTimeSeries(url string) ([]byte, *common.ONSError)
}
