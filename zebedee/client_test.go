package zebedee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testClient struct{}

func (*testClient) Get(url string) (resp *http.Response, err error) {
	return httptest.NewRecorder().Result(), nil
}

func TestGetData(t *testing.T) {

	Client := &testClient{}
	//Url = "http://localhost:8082"

	resp, err := Client.Get("/")

	if err != nil {
		t.Error("Error", err)
	}

	if resp == nil {
		t.Error("Nil response")
	}
}
