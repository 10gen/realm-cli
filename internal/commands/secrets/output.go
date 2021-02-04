package secrets

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	headerID      = "ID"
	headerName    = "Name"
	headerDeleted = "Deleted"
	headerDetails = "Details"
)

type secretOutputs []secretOutput

type secretOutput struct {
	secret realm.Secret
	err error
}

type secretTableRowModifier func(secretOutput, map[string]interface{})

func secretOutputComparerBySuccess(outputs secretOutputs) func(i, j int) bool{
	return func(i, j int) bool {
		return outputs[i].err != nil && outputs[j].err == nil
	}
}

func secretHeaders() []string{
	return []string {headerID, headerName, headerDetails}
}

func secretTableRows(outputs secretOutputs, modifier secretTableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, secretTableRow(output, modifier))
	}
	return rows
}

func secretTableRow(output secretOutput, modifier secretTableRowModifier) map[string]interface{} {
	row := map[string]interface{} {
		headerID: output.secret.ID,
		headerName: output.secret.Name,
	}
	modifier(output, row)
	return row
}


func displaySecretOption(secret realm.Secret) string {
	return secret.ID + terminal.DelimiterInline + secret.Name
}
