package content

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"encoding/json"
	zebedeeModel "github.com/ONSdigital/dp-content-resolver/zebedee/model"
	"errors"
	"github.com/ONSdigital/dp-frontend-renderer/model"
	"reflect"
)

type testExpectations struct {
	taxonomy []model.TaxonomyNode
	err error
	zebedeeBytes []byte
}

var expectedResults testExpectations

func TestGetTaxonomy(t *testing.T) {
	GetTaxonomy = errorResponse

	Convey("Should return empty result & error if error getting taxonomy data from zebedee.", t, func() {
		expectedResults = testExpectations{
			taxonomy: make([]model.TaxonomyNode, 0),
			err: errors.New("Taxonomy error."),
			zebedeeBytes: make([]byte, 0),
		}
		taxonomyList, err := resolveTaxonomy("/")
		So(reflect.DeepEqual(taxonomyList, expectedResults.taxonomy), ShouldBeTrue)
		So(err, ShouldEqual, expectedResults.err)
	})
}

func validTaxonomy(url string, depth int) ([]byte, error) {
	description := zebedeeModel.PageDescription{Title: "Hello World"}

	taxonomyChildren := make([]zebedeeModel.Taxonomy, 0)
	taxonomyChildren = append(taxonomyChildren, zebedeeModel.Taxonomy{
		URI: "/child1/child2",
		Description: zebedeeModel.PageDescription{Title: "Child Hello World"},
	})

	taxonomy := &zebedeeModel.Taxonomy{
		URI: "/child1",
		Description: description,
		Children: taxonomyChildren,
	}
	return json.Marshal(taxonomy)
}

func errorResponse(url string, depth int) ([]byte, error) {
	return expectedResults.zebedeeBytes, expectedResults.err
}



