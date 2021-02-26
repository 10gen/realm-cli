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
	err    error
}

type tableRowModifier func(secretOutput, map[string]interface{})

func tableHeaders(additionalHeaders ...string) []string {
	return append([]string{headerID, headerName}, additionalHeaders...)
}

func tableRows(outputs secretOutputs, modifier tableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, tableRow(output, modifier))
	}
	return rows
}

func tableRow(output secretOutput, modifier tableRowModifier) map[string]interface{} {
	row := map[string]interface{}{
		headerID:   output.secret.ID,
		headerName: output.secret.Name,
	}
	modifier(output, row)
	return row
}

func displaySecretOption(secret realm.Secret) string {
	return secret.ID + terminal.DelimiterInline + secret.Name
}
