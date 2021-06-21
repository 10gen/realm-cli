package user

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

const (
	headerEmail                  = "Email"
	headerEnabled                = "Enabled"
	headerID                     = "ID"
	headerLastAuthenticationDate = "Last Authenticated"
	headerName                   = "Name"
	headerType                   = "Type"
	headerDeleted                = "Deleted"
	headerDetails                = "Details"
	headerRevoked                = "Session Revoked"
)

type newUserOutputs struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type newUserAPIKeyOutputs struct {
	newUserOutputs
	Name string `json:"name"`
	Key  string `json:"key"`
}

type newUserEmailOutputs struct {
	newUserOutputs
	Email interface{} `json:"email"`
	Type  string      `json:"type"`
}

type userOutputs []userOutput

func (outputs userOutputs) byProviderType() map[realm.AuthProviderType]userOutputs {
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

type tableRowModifier func(userOutput, map[string]interface{})

func tableHeaders(authProviderType realm.AuthProviderType) []string {
	var tableHeaders []string

	switch authProviderType {
	case realm.AuthProviderTypeAPIKey:
		tableHeaders = append(tableHeaders, headerName)
	case realm.AuthProviderTypeUserPassword:
		tableHeaders = append(tableHeaders, headerEmail)
	}

	return append(tableHeaders, headerID, headerType)
}

func tableRows(authProviderType realm.AuthProviderType, outputs userOutputs, tableRowModifier tableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, tableRow(authProviderType, output, tableRowModifier))
	}
	return rows
}

func tableRow(authProviderType realm.AuthProviderType, output userOutput, tableRowModifier tableRowModifier) map[string]interface{} {
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
