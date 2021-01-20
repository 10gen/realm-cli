package user

import "github.com/10gen/realm-cli/internal/cloud/realm"

const (
	headerAPIKey                 = "API Key"
	headerEmail                  = "Email"
	headerEnabled                = "Enabled"
	headerID                     = "ID"
	headerLastAuthenticationDate = "Last Authenticated"
	headerName                   = "Name"
	headerType                   = "Type"
	headerDeleted                = "Deleted"
	headerDetails                = "Details"
)

type userOutput struct {
	user realm.User
	err  error
}

func getUserOutputComparerBySuccess(outputs []userOutput) func(i, j int) bool {
	return func(i, j int) bool {
		return !(outputs[i].err != nil && outputs[i].err == nil)
	}
}
