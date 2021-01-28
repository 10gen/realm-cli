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
	headerRevoked                = "Revoked"
)

type userOutputs []userOutput

func (outputs userOutputs) mapByProviderType() map[realm.AuthProviderType]userOutputs {
	var outputsM = map[realm.AuthProviderType]userOutputs{}
	for _, output := range outputs {
		for _, identity := range output.user.Identities {
			outputsM[identity.ProviderType] = append(outputsM[identity.ProviderType], output)
		}
	}
	return outputsM
}

type userOutput struct {
	user realm.User
	err  error
}

func getUserOutputComparerBySuccess(outputs userOutputs) func(i, j int) bool {
	return func(i, j int) bool {
		return outputs[i].err != nil && outputs[j].err == nil
	}
}

type userTableRowModifier func(userOutput, map[string]interface{})

func userTableHeaders(authProviderType realm.AuthProviderType) []string {
	var headers []string

	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		headers = append(headers, headerName)
	case realm.AuthProviderTypeUserPassword:
		headers = append(headers, headerEmail)
	}

	return append(headers, headerID, headerType)
}

func userTableRows(authProviderType realm.AuthProviderType, outputs userOutputs, tableRowModifier userTableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, userTableRow(authProviderType, output, tableRowModifier))
	}
	return rows
}

func userTableRow(authProviderType realm.AuthProviderType, output userOutput, tableRowModifier userTableRowModifier) map[string]interface{} {
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

	tableRowModifier(output, row)
	return row
}
