package secrets

// Secret represents an app secret
type Secret struct {
	ID    string `json:"_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
