package zebedee

import (
	"bytes"
	"errors"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const requestContextID = "1234"

// Set these with the values you want to be returned for your test case.
var responseStub *http.Response
var errorStub error
var dataStub []byte
var pageTypeStub string
var onsErrorStub *common.ONSError
var responseBodyReadErrStub error
var responseBodyBytesStub []byte

// Struct replaces client in target code - allowing you to return stub data.
type testClient struct{}

func (*testClient) Get(url string) (resp *http.Response, err error) {
	recorder := httptest.NewRecorder()
	recorder.Header().Add("ONS-Page-Type", "home_page")
	return recorder.Result(), nil
}

// Stub Implementation of the Do func - set returned vars with the values required by your test case.
func (*testClient) Do(req *http.Request) (*http.Response, error) {
	return responseStub, errorStub
}

func ReadBodyMock(io.Reader) ([]byte, error) {
	return responseBodyBytesStub, responseBodyReadErrStub
}

func TestGetData(t *testing.T) {
	// create stub http client for test
	testHTTPClient := &testClient{}

	// inject it into an instance of zebedeeHTTPClient
	zebedeeClient := Client{testHTTPClient, "http://zebedeeUri"}

	Convey("Should return empty data, page type and correct error if zebedee.get data fails.", t, func() {

		// Set stub data to return for this test case.
		errorStub = errors.New("Zebedee get data error")
		dataStub = make([]byte, 0)
		pageTypeStub = ""
		onsErrorStub = common.NewONSError(errorStub, zebedeeGetError)
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)
		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		ShouldEqual(data, dataStub)
		ShouldEqual(pageType, pageTypeStub)
		ShouldEqual(err, onsErrorStub)
	})

	Convey("Should return empty data & page type and appropriate error if Zebedee returns an unexpected status code.", t, func() {

		// Set stub data to return for this test case.
		onsErrorStub = &common.ONSError{RootError: errors.New("Unexpected Response status code")}
		onsErrorStub = common.NewONSError(errors.New("Unexpected Response status code"), incorrectStatusCodeErrDesc)
		onsErrorStub.AddParameter("zebedeeURI", "/data")
		onsErrorStub.AddParameter("query", "/")
		onsErrorStub.AddParameter("expectedStatusCode", 200)
		onsErrorStub.AddParameter("actualStatusCode", 500)
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)
		errorStub = nil
		responseStub = &http.Response{StatusCode: 500}
		dataStub = make([]byte, 0)
		pageTypeStub = ""

		// Run test
		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		// assert results.
		So(err, ShouldResemble, onsErrorStub)
		So(0, ShouldEqual, bytes.Compare(data, dataStub))
		So(pageType, ShouldEqual, pageTypeStub)
	})

	Convey("Should return empty data & pageType & appropriate error if there is an error reading the response body.", t, func() {

		zebedeeClient.setResponseReader(ReadBodyMock)

		dataStub = []byte("")
		rootErr := errors.New("it broked")
		pageTypeStub = ""
		onsErrorStub = common.NewONSError(rootErr, "error reading response body")
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)

		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, "home_page")
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(""))
		responseBodyReadErrStub = rootErr
		responseBodyBytesStub = []byte("")

		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		So(err, ShouldResemble, onsErrorStub)
		So(data, ShouldResemble, dataStub)
		So(pageType, ShouldEqual, pageTypeStub)
	})

	Convey("Should return expected data, pageType for successful calls.", t, func() {

		zebedeeClient.setResponseReader(ReadBodyMock)

		body := "I am Success!"

		dataStub = []byte(body)
		pageTypeStub = "home_page"
		onsErrorStub = nil

		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, "home_page")
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(body))
		responseBodyReadErrStub = nil
		responseBodyBytesStub = []byte(body)

		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		So(err, ShouldResemble, onsErrorStub)
		So(data, ShouldResemble, dataStub)
		So(pageType, ShouldEqual, pageTypeStub)
	})
}
