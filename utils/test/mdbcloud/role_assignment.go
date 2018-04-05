package mdbcloud

// RoleAssignment represents the role that entity has with respect
// to a group
type RoleAssignment struct {
	GroupID  string `json:"groupId"`
	RoleName string `json:"roleName"`
}
