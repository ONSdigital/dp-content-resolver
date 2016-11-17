package zebedee

import (
    "github.com/ONSdigital/go-ns/log"
    "io/ioutil"
    "net/http"
    "strconv"
    "time"
    "errors"
    "github.com/ONSdigital/go-ns/common"
    "io"
)

const uriParam = "uri"
const dataApi = "/data"
const taxonomyApi = "/taxonomy"
const breadcrumbApi = "/parents"
const pageTypeHeader = "Ons-Page-Type"
const zebedeeGetError = "GET zebedee/data request returned an unexpected error."

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

// Hide read response body behind behind type to allow us to replace with stub during tests.
type responseBodyReader func(io.Reader) ([]byte, error)

var resReader responseBodyReader = ioutil.ReadAll

// CreateClient will create a new ZebedeeHTTPClient for the given url and timeout.
func CreateClient(timeout time.Duration, zebedeeURL string) *Client {
    return &Client{
        &http.Client{
            Timeout: timeout,
        },
        zebedeeURL}
}

// GetData will call Zebedee and return the data it provides in a []byte
func (zebedee *Client) GetData(uri string) (data []byte, pageType string, err *common.ONSError) {
    var response *http.Response

    request, error := zebedee.buildGetRequest(dataApi, []parameter{{name: uriParam, value: uri}})
    if err != nil {
        return data, pageType, common.NewONSError(error, "error creating zebedee request.")
    }

    // TODO add request id to header.
    response, error = zebedee.httpClient.Do(request)

    if error != nil {
        return data, pageType, common.NewONSError(error, zebedeeGetError)
    }

    if response.StatusCode != 200 {
        error := &common.ONSError{RootError: errors.New("Unexpected Response status code")};
        error.AddParameter("zebedeeURI", request.URL.Path)
        error.AddParameter("expectedStatusCode", 200)
        error.AddParameter("actualStatusCode", response.StatusCode)
        error.AddParameter("query", request.URL.Query().Get("uri"))
        return data, pageType, error
    }

    data, error = resReader(response.Body)
    defer response.Body.Close()

    if error != nil {
        return data, pageType, common.NewONSError(error, "error reading response body")
    }

    pageType = response.Header.Get(pageTypeHeader)
    log.Debug("Identified page type", log.Data{"page type": pageType})
    return
}

// GetTaxonomy gets the taxonomy structure of the website from Zebedee
func (zebedee *Client) GetTaxonomy(uri string, depth int) ([]byte, *common.ONSError) {
    return zebedee.get(taxonomyApi, []parameter{{name: uriParam, value: uri}, {name: "depth", value: strconv.Itoa(depth)}})
}

func (zebedee *Client) GetParents(uri string) ([]byte, *common.ONSError) {
    return zebedee.get(breadcrumbApi, []parameter{{name: uriParam, value: uri}})
}

func (zebedee *Client) GetTimeSeries(uri string) ([]byte, *common.ONSError) {
    return zebedee.get(dataApi, []parameter{{name: uriParam, value: uri}, {name: "series"}})
}

func (zebedee *Client) get(path string, params []parameter) ([]byte, *common.ONSError) {
    request, err := zebedee.buildGetRequest(path, params)
    if err != nil {
        return nil, common.NewONSError(err, "error creating zebedee request")
    }

    response, err := zebedee.httpClient.Do(request)
    defer response.Body.Close()

    if err != nil {
        return nil, common.NewONSError(err, "error performing zebedee request")
    }

    if response.StatusCode != 200 {
        onsError := &common.ONSError{RootError: errors.New("Unexpected error response status")}
        onsError.AddParameter("expectedStatusCode", 200)
        onsError.AddParameter("actualStatusCode", response.StatusCode)
        onsError.AddParameter("zebedeeURI", request.URL.Path)
        return nil, onsError
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, common.NewONSError(err, "error reading zebedee response body")
    }
    return body, nil
}

// TODO add request header with request id here.
func (zebedee *Client) buildGetRequest(url string, params []parameter) (*http.Request, error) {
    request, err := http.NewRequest("GET", zebedee.url + url, nil)
    if err != nil {
        return nil, err
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

func (zebedee *Client) setResponseReader(f func(io.Reader) ([]byte, error)) {
    resReader = f
}
