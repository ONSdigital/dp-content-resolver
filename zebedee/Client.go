package zebedee

import (
	"errors"
	"fmt"
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

type parameter struct {
	name  string
	value string
}

// GetData will call Zebedee and return the data it provides in a []byte
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

func (zebedee *zebedeeHTTPClient) GetTaxonomy(url string) ([]byte, error) {
	params := []parameter{{name: "uri", value: url}}
	taxonomy, _ := zebedee.get("/taxonomy", params)
	fmt.Printf("Taxonomy \n%v\n", string(taxonomy))
	return taxonomy, nil
}

func (zebedee *zebedeeHTTPClient) get(path string, params []parameter) ([]byte, error) {
	request, err := zebedee.buildGetRequest(path, params)
	if err != nil {
		log.Error(err, log.Data{"message": "error creating zebedee request"})
		return nil, nil
	}

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Error(err, log.Data{"message": "Status code not 200"})
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err, log.Data{"message": "failed to read response body"})
		return nil, err
	}
	return body, nil
}

func (zebedee *zebedeeHTTPClient) buildGetRequest(url string, params []parameter) (*http.Request, error) {
	request, err := http.NewRequest("GET", zebedee.url+url, nil)
	if err != nil {
		log.Error(err, log.Data{"message": "error creating zebedee request"})
		return nil, nil
	}

	if len(params) > 0 {
		query := request.URL.Query()
		for _, param := range params {
			query.Add(param.name, param.value)
		}
		request.URL.RawQuery = query.Encode()
	}
	return request, nil
}
