package zebedee

import (
	"errors"
	"github.com/ONSdigital/go-ns/log"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Client provides the interface to zebedee.
type Client interface {
	GetData(url string) (data []byte, pageType string, err error)
}

// httpClient provides only the methods of http.client that we are using allowing it to be mocked.
type httpClient interface {
	Get(url string) (resp *http.Response, err error)
}

// zebedeeHTTPClient is a http specific implementation of zebedeeClient, using our httpClient interface
type zebedeeHTTPClient struct {
	httpClient httpClient
	url        string
}

// GetData is implemented on zebedeeHTTPClient to satisfy the zebedeeClient interface.
func (zebedee *zebedeeHTTPClient) GetData(url string) (data []byte, pageType string, err error) {
	// get page data from zebedee (homepage)
	log.Debug("Getting data from zebedee", log.Data{"url": url})
	var response *http.Response
	response, err = zebedee.httpClient.Get(zebedee.url + "/data?uri=" + url)
	if err != nil {
		log.Debug("Failed to get data from zebedee", log.Data{"url": url})
		return
	}

	// check response codes
	if response.StatusCode != 200 {
		// read the response body to ensure its memory is freed.
		io.Copy(ioutil.Discard, response.Body)
		err = ErrUnexpectedStatusCode
		return
	}

	// unmarshal into homepage object
	data, err = ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Debug("Failed to deserialise json from zebedee", log.Data{"url": url})
		return
	}

	pageType = response.Header.Get("ONS-Page-Type") //"home_page"
	log.Debug("Identified page type", log.Data{"page type": pageType})
	return
}

// ErrUnexpectedStatusCode is the error returned when you get an unexpected error code.
var ErrUnexpectedStatusCode = errors.New("Unexpected status code")

// CreateClient will create a new zebedeeHTTPClient for the given url and timeout.
func CreateClient(timeout time.Duration, zebedeeURL string) Client {
	return &zebedeeHTTPClient{
		&http.Client{
			Timeout: timeout,
		},
		zebedeeURL}
}
