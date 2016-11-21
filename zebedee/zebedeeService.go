package zebedee

// Defines interface of zebedee service.
type Service interface {
	GetData(url string) (data []byte, pageType string, err error)
	GetTaxonomy(url string, depth int) ([]byte, error)
	GetParents(url string) ([]byte, error)
	GetTimeSeries(url string) ([]byte, error)
}
