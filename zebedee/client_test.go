package zebedee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testClient struct{}

func (*testClient) Get(url string) (resp *http.Response, err error) {
	recorder := httptest.NewRecorder()
	recorder.Header().Add("ONS-Page-Type", "home_page")
	return recorder.Result(), nil
}

func (*testClient) Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestGetData(t *testing.T) {

	// create stub http client for test
	client := &testClient{}

	// inject it into an instance of zebedeeHTTPClient
	zebedeeClient := zebedeeHTTPClient{client, "http://zebedeeUri"}

	data, pageType, err := zebedeeClient.GetData("/")

	if err != nil {
		t.Error("Error", err)
	}

	if data == nil {
		t.Error("Data should not be nil")
	}

	if len(pageType) == 0 {
		t.Error("Pagetype should have a value")
	}

	if pageType != "home_page" {
		t.Error("Expected a pageType value of home_page but was", pageType)
	}
}
