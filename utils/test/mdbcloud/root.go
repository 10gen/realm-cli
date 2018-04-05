package mdbcloud

// Root is the view returned by the API when using the base URL
type Root struct {
	Links []Link `json:"links"`
}
