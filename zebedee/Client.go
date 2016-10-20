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

type parameter struct {
	name  string
	value string
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

func GetTaxonomy(url string) ([]byte, error) {
	params := []parameter{{name: "uri", value: url}}
	taxonomy, _ := zebedeeGet("/taxonomy", params)
	fmt.Printf("Taxonomy \n%v\n", string(taxonomy))
	return taxonomy, nil
}

func zebedeeGet(path string, params []parameter) ([]byte, error) {
	request, err := buildGetRequest(path, params)
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

func buildGetRequest(url string, params []parameter) (*http.Request, error) {
	request, err := http.NewRequest("GET", defaultClient.zebedeeURL+url, nil)
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
