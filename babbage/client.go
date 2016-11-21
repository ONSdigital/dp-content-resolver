package babbage

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrCreatingRequest      = errors.New("error creating request")
	ErrCallingBabbage       = errors.New("error calling babbage")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrReadingBody          = errors.New("error reading body")
)

type babbageError struct {
	error
	BabbageError error
	URI          string
	StatusCode   int
}

type bClient struct {
	*http.Client
	URL string
}

func CreateClient(timeout time.Duration, babbageURL string) Service {
	return &bClient{
		Client: &http.Client{
			Timeout: timeout,
		},
		URL: babbageURL,
	}
}

func (babbage *bClient) GetTimeSeries(uri string) ([]byte, error) {
	b, _, err := babbage.get(babbage.URL+uri+"/data", nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (babbage *bClient) get(uri string, query *url.Values) ([]byte, *http.Response, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, babbageError{err, ErrCreatingRequest, uri, 0}
	}

	res, err := babbage.Client.Do(req)
	if err != nil {
		return nil, res, babbageError{err, ErrCallingBabbage, uri, 0}
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		io.Copy(ioutil.Discard, res.Body)
		return nil, res, babbageError{err, ErrUnexpectedStatusCode, uri, res.StatusCode}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res, babbageError{err, ErrReadingBody, uri, res.StatusCode}
	}

	return b, res, nil
}
