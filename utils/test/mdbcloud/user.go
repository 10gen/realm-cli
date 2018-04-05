package mdbcloud

// User represents a Cloud Manager user
type User struct {
	ID           string           `json:"id"`
	FirstName    string           `json:"firstName"`
	LastName     string           `json:"lastName"`
	EmailAddress string           `json:"emailAddress"`
	Roles        []RoleAssignment `json:"roles"`
	Links        []Link           `json:"links"`
}
