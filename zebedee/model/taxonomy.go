package model

import (
	"github.com/ONSdigital/dp-frontend-renderer/model"
)

type Taxonomy struct {
	URI         string         `json:"uri"`
	Description PageDescription `json:"description"`
	Children    []Taxonomy `json:"children,omitempty"`
}

// Recursively convert a list zebedee taxonomy nodes to a list of renderer taxonomy nodes.
func (zebedeeTaxonomy *Taxonomy) Map() model.TaxonomyNode {
	var taxonomyNode model.TaxonomyNode

	taxonomyNode.Title = zebedeeTaxonomy.Description.Title
	taxonomyNode.URI = zebedeeTaxonomy.URI

	if len(zebedeeTaxonomy.Children) > 0 {
		var children = make([]model.TaxonomyNode, 0)
		for _, zebedeeTaxonomyNode := range zebedeeTaxonomy.Children {
			children = append(children, zebedeeTaxonomyNode.Map())
		}
		taxonomyNode.Children = children
	}
	return taxonomyNode
}