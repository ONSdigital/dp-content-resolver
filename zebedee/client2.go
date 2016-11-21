package zebedee

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrCreatingRequest      = errors.New("error creating request")
	ErrCallingZebedee       = errors.New("error calling zebedee")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrReadingBody          = errors.New("error reading body")
)

type zebedeeError struct {
	error
	ZebedeeError error
	URI          string
	StatusCode   int
}

type zClient struct {
	*http.Client
	URL string
}

func CreateZClient(timeout time.Duration, zebedeeURL string) Service {
	return &zClient{
		Client: &http.Client{
			Timeout: timeout,
		},
		URL: zebedeeURL,
	}
}

func (z *zClient) GetData(uri string) ([]byte, string, error) {
	q := &url.Values{}
	q.Add("uri", uri)

	b, res, err := z.get(z.URL+"/data", q)
	if err != nil {
		return nil, "", err
	}

	return b, res.Header.Get("ONS-Page-Type"), nil
}

func (z *zClient) GetTaxonomy(uri string, depth int) ([]byte, error) {
	q := &url.Values{}
	q.Add("uri", uri)
	q.Add("depth", fmt.Sprintf("%d", depth))

	b, _, err := z.get(z.URL+"/taxonomy", q)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (z *zClient) GetParents(uri string) ([]byte, error) {
	q := &url.Values{}
	q.Add("uri", uri)

	b, _, err := z.get(z.URL+"/parents", q)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (z *zClient) GetTimeSeries(uri string) ([]byte, error) {
	q := &url.Values{}
	q.Add("uri", uri)
	q.Add("series", "1")

	b, _, err := z.get(z.URL+"/data", q)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (z *zClient) get(uri string, query *url.Values) ([]byte, *http.Response, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, zebedeeError{err, ErrCreatingRequest, uri, 0}
	}

	res, err := z.Client.Do(req)
	if err != nil {
		return nil, res, zebedeeError{err, ErrCallingZebedee, uri, 0}
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		io.Copy(ioutil.Discard, res.Body)
		return nil, res, zebedeeError{err, ErrUnexpectedStatusCode, uri, res.StatusCode}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res, zebedeeError{err, ErrReadingBody, uri, res.StatusCode}
	}

	return b, res, nil
}
