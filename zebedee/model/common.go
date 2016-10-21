package model

// PageDescription is a common section for every page containing common fields.
type PageDescription struct {
	Title    string `json:"title"`
	Summary  string `json:"description"`
	Keywords []string
}
