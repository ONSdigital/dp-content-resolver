package zebedee

import (
	"errors"
	"github.com/ONSdigital/go-ns/log"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// httpClient provides only the methods of http.client that we are using allowing it to be mocked.
type httpClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

// Client holds the required fields to call Zebedee.
type Client struct {
	httpClient httpClient
	url        string
}

type parameter struct {
	name  string
	value string
}

// GetData will call Zebedee and return the data it provides in a []byte
func (zebedee *Client) GetData(url string) (data []byte, pageType string, err error) {
	// get page data from zebedee (homepage)
	log.Debug("Getting data from zebedee", log.Data{"url": url})
	var response *http.Response
	response, err = zebedee.httpClient.Get(zebedee.url + "/data?uri=" + url)
	if err != nil {
		log.Error(err, log.Data{"url": url})
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
		log.Error(err, log.Data{"url": url})
		return
	}

	pageType = response.Header.Get("ONS-Page-Type") //"home_page"
	log.Debug("Identified page type", log.Data{"page type": pageType})
	return
}

// ErrUnexpectedStatusCode is the error returned when you get an unexpected error code.
var ErrUnexpectedStatusCode = errors.New("Unexpected status code")

// CreateClient will create a new ZebedeeHTTPClient for the
// given url and timeout.
func CreateClient(timeout time.Duration, zebedeeURL string) *Client {
	return &Client{
		&http.Client{
			Timeout: timeout,
		},
		zebedeeURL}
}

// GetTaxonomy gets the taxonomy structure of the website from Zebedee
func (zebedee *Client) GetTaxonomy(url string, depth int) ([]byte, error) {
	return zebedee.get("/taxonomy", []parameter{{name: "uri", value: url}, {name: "depth", value: strconv.Itoa(depth)}})
}

func (zebedee *Client) GetParents(url string) ([]byte, error) {
	return zebedee.get("/parents", []parameter{{name: "uri", value: url}})
}

func (zebedee *Client) get(path string, params []parameter) ([]byte, error) {
	request, err := zebedee.buildGetRequest(path, params)
	if err != nil {
		log.Error(err, log.Data{"message": "error creating zebedee request"})
		return nil, nil
	}

	response, err := zebedee.httpClient.Do(request)
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

func (zebedee *Client) buildGetRequest(url string, params []parameter) (*http.Request, error) {
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
