package zebedee

import (
	"errors"
	"github.com/ONSdigital/go-ns/log"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var defaultClient *client

// ErrUnexpectedStatusCode is the error returned when you get an unexpected error code.
var ErrUnexpectedStatusCode = errors.New("Unexpected status code")

type client struct {
	*http.Client
	zebedeeURL string
}

// Init will set up the zebedee client with the given timeout duration and base Zebedee URL.
func Init(timeout time.Duration, zebedeeURL string) {
	defaultClient = &client{&http.Client{
		Timeout: timeout,
	}, zebedeeURL,
	}
}

// GetData will call Zebedee and return the data it provides in a []byte
func GetData(url string) (data []byte, pageType string, err error) {

	// get page data from zebedee (homepage)
	log.Debug("Getting data from zebedee", log.Data{"url": url})
	var response *http.Response
	response, err = defaultClient.Get(defaultClient.zebedeeURL + "/data?uri=" + url)
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
