package homePage

//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
//	rendererModel "github.com/ONSdigital/dp-frontend-renderer/model"
//	. "github.com/smartystreets/goconvey/convey"
//	"reflect"
//	"testing"
//)
//
//type testExpectations struct {
//	taxonomy     []rendererModel.TaxonomyNode
//	err          error
//	zebedeeBytes []byte
//	pageType     string
//}
//
//type zebedeeServiceMock struct {
//	testValues testExpectations
//}
//
//// Mock behaviour for get taxonomy
//func (mock *zebedeeServiceMock) GetTaxonomy(uri string, depth int) ([]byte, error) {
//	return mock.testValues.zebedeeBytes, mock.testValues.err
//}
//
//// Mock behaviour for get data.
//func (mock *zebedeeServiceMock) GetData(uri string) ([]byte, string, error) {
//	return mock.testValues.zebedeeBytes, mock.testValues.pageType, mock.testValues.err
//}
//
//// Mock behaviour for get parents.
//func (mock *zebedeeServiceMock) GetParents(uri string) ([]byte, error) {
//	return mock.testValues.zebedeeBytes, mock.testValues.err
//}
//
//var expectedResults testExpectations
//
//func TestGetTaxonomy(t *testing.T) {
//
//	var rList []rendererModel.TaxonomyNode
//	var byteList []byte
//
//	Convey("Should return empty result & error if error getting taxonomy data from zebedee.", t, func() {
//		//		GetTaxonomy = errorResponse
//
//		expectedResults = testExpectations{
//			taxonomy:     rList,
//			err:          errors.New("Taxonomy error."),
//			zebedeeBytes: byteList,
//		}
//
//		ZebedeeService = &zebedeeServiceMock{testValues: expectedResults}
//
//		taxonomyList, err := resolveTaxonomy("/")
//		So(reflect.DeepEqual(taxonomyList, expectedResults.taxonomy), ShouldBeTrue)
//		So(err, ShouldEqual, expectedResults.err)
//	})
//
//	Convey("Should return empty result & error for error unmarshalling zebedee taxonomy json.", t, func() {
//		//GetTaxonomy = unmarshallError
//		expectedResults = testExpectations{
//			taxonomy:     rList,
//			err:          fmtError("resolveTaxonomy", "Error unmarshalling taxonomy json."),
//			zebedeeBytes: byteList,
//		}
//
//		ZebedeeService = &zebedeeServiceMock{testValues: expectedResults}
//
//		taxonomyList, err := resolveTaxonomy("/")
//		So(reflect.DeepEqual(taxonomyList, expectedResults.taxonomy), ShouldBeTrue)
//		So(err, ShouldNotBeNil)
//		So(err.Error(), ShouldEqual, expectedResults.err.Error())
//	})
//
//	Convey("Should return the a list of TaxonomyNode for a successful zebedee response", t, func() {
//
//		zebedeeTaxonomyList, rendererTaxonomyList := getTaxonomyModels()
//		zebedeeTaxonomyJson, _ := json.Marshal(zebedeeTaxonomyList)
//
//		expectedResults = testExpectations{
//			taxonomy:     rendererTaxonomyList,
//			err:          nil,
//			zebedeeBytes: zebedeeTaxonomyJson,
//		}
//
//		ZebedeeService = &zebedeeServiceMock{testValues: expectedResults}
//
//		//GetTaxonomy = validTaxonomy
//		actualTaxonomy, err := resolveTaxonomy("/")
//		So(err, ShouldBeNil)
//
//		fmt.Printf("\nactual %v\n", actualTaxonomy)
//		fmt.Printf("\nexpected %v\n", rendererTaxonomyList)
//
//		So(reflect.DeepEqual(actualTaxonomy, rendererTaxonomyList), ShouldBeTrue)
//	})
//}
//
//// Create the expected zebedee & renderer taxonomy models.
//func getTaxonomyModels() ([]zebedeeModel.ContentNode, []rendererModel.TaxonomyNode) {
//	uri := "/uri"
//	childUri := uri + "/child"
//	title := "Hello World"
//	var zebedeeList []zebedeeModel.ContentNode
//	var rendererList []rendererModel.TaxonomyNode
//
//	// Create a child and a parent of the Zebedee Taxonomy Model.
//	zebedeeTaxonomyChild := zebedeeModel.ContentNode{
//		URI:         childUri,
//		Description: zebedeeModel.PageDescription{Title: title},
//	}
//	zebedeeModelTaxonomy := zebedeeModel.ContentNode{
//		URI:         childUri,
//		Description: zebedeeModel.PageDescription{Title: title},
//		Children:    []zebedeeModel.ContentNode{zebedeeTaxonomyChild},
//	}
//	zebedeeList = append(zebedeeList, zebedeeModelTaxonomy)
//
//	// Create a child and a parent of the Renderer Taxonomy Model.
//	rendererTaxonomyChild := rendererModel.TaxonomyNode{
//		URI:   childUri,
//		Title: title,
//	}
//	renderModelTaxonomy := rendererModel.TaxonomyNode{
//		URI:      childUri,
//		Title:    title,
//		Children: []rendererModel.TaxonomyNode{rendererTaxonomyChild},
//	}
//	rendererList = append(rendererList, renderModelTaxonomy)
//
//	return zebedeeList, rendererList
//}
