package user

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

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
		return outputs[i].err != nil && outputs[j].err == nil
	}
}

func userTableHeaders(authProviderType realm.AuthProviderType) []string {
	var headers []string
	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		headers = append(headers, headerName)
	case realm.AuthProviderTypeUserPassword:
		headers = append(headers, headerEmail)
	}
	headers = append(
		headers,
		headerID,
		headerType,
	)
	return headers
}

func userTableRows(authProviderType realm.AuthProviderType, outputs []userOutput, additionalRowOutput func(userOutput, map[string]interface{})) []map[string]interface{} {
	userTableRows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		userTableRows = append(userTableRows, userTableRow(authProviderType, output, additionalRowOutput))
	}
	return userTableRows
}

func userTableRow(authProviderType realm.AuthProviderType, output userOutput, additionalRowOutput func(userOutput, map[string]interface{})) map[string]interface{} {
	row := map[string]interface{}{
		headerID:   output.user.ID,
		headerType: output.user.Type,
	}
	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		row[headerName] = output.user.Data[userDataName]
	case realm.AuthProviderTypeUserPassword:
		row[headerEmail] = output.user.Data[userDataEmail]
	}
	additionalRowOutput(output, row)
	return row
}
