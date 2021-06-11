package accesslist

import "github.com/10gen/realm-cli/internal/cloud/realm"

const (
	headerIP      = "IP Address"
	headerComment = "Comment"
	headerDeleted = "Deleted"
	headerDetails = "Details"
)

var (
	tableHeaders = []string{headerIP, headerComment}
)

type allowedIPOutputs []allowedIPOutput

type allowedIPOutput struct {
	allowedIP realm.AllowedIP
	err       error
}

type tableRowModifier func(allowedIPOutput, map[string]interface{})

func tableRows(outputs allowedIPOutputs, modifier tableRowModifier) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(outputs))
	for _, output := range outputs {
		rows = append(rows, tableRow(output, modifier))
	}
	return rows
}

func tableRow(output allowedIPOutput, modifier tableRowModifier) map[string]interface{} {
	row := map[string]interface{}{
		headerIP:      output.allowedIP.ID,
		headerComment: output.allowedIP.Comment,
	}
	modifier(output, row)
	return row
}
