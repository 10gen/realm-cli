package mdbcloud

// AtlasDBUser represents an Atlas Database User
type AtlasDBUser struct {
	DatabaseName string            `json:"databaseName"`
	Roles        []AtlasDBUserRole `json:"roles"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
}

// AtlasDBUserRole represents an Atlas Database User's role
type AtlasDBUserRole struct {
	DatabaseName string `json:"databaseName"`
	RoleName     string `json:"roleName"`
}

// The set of known Atlas Database User roles
const (
	AtlasDBRoleReadWriteAnyDatabase = "readWriteAnyDatabase"
)
