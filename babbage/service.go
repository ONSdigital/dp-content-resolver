package babbage

// Defines interface of babbage service.
type Service interface {
	GetTimeSeries(url string) ([]byte, error)
}
