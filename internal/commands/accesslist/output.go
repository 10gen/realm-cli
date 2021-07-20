package accesslist

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	headerAddress = "Address"
	headerComment = "Name"
	headerDeleted = "Deleted"
	headerDetails = "Details"
)

type deleteAllowedIPOutputs []deleteAllowedIPOutput

type deleteAllowedIPOutput struct {
	allowedIP realm.AllowedIP
	err       error
}

type tableRowModifier func(deleteAllowedIPOutput, map[string]interface{})

func tableHeaders(additionalHeaders ...string) []string {
	return append([]string{headerAddress, headerComment}, additionalHeaders...)
}

func tableRows(outputs deleteAllowedIPOutputs, modifier tableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, tableRow(output, modifier))
	}
	return rows
}

func tableRow(output deleteAllowedIPOutput, modifier tableRowModifier) map[string]interface{} {
	row := map[string]interface{}{
		headerAddress: output.allowedIP.Address,
		headerComment: output.allowedIP.Comment,
	}
	modifier(output, row)
	return row
}

func displayAllowedIPOption(allowedIP realm.AllowedIP) string {
	option := allowedIP.ID + terminal.DelimiterInline + allowedIP.Address
	if allowedIP.Comment == "" {
		return option
	}
	return option + terminal.DelimiterInline + allowedIP.Comment
}
