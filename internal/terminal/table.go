package terminal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	logFieldHeaders = "headers"
	logFieldData    = "data"
)

// set of exported spacing options
const (
	Indent = "  "
	Gutter = "  "
)

var (
	tableFields = []string{logFieldMessage, logFieldData, logFieldHeaders}
)

type table struct {
	message      string
	headers      []string
	data         []map[string]string
	columnWidths map[string]int
}

func newTable(message string, headers []string, data []map[string]interface{}) table {
	var t table

	if len(headers) == 0 {
		return t
	}

	t.message = message
	t.headers = headers
	t.data = make([]map[string]string, 0, len(data))
	t.columnWidths = make(map[string]int, len(headers))

	for _, header := range headers {
		t.columnWidths[header] = len(header)
	}

	for _, row := range data {
		if len(row) == 0 {
			continue
		}
		r := make(map[string]string)
		for _, header := range t.headers {
			value := parseValue(row[header])
			if width := len(value); width > t.columnWidths[header] {
				t.columnWidths[header] = width
			}
			r[header] = value
		}
		t.data = append(t.data, r)
	}
	return t
}

func (t table) Message() (string, error) {
	if err := t.validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf(`%s
%s
%s
%s`, t.message, t.headerString(), t.dividerString(), t.dataString()), nil
}

func (t table) Payload() ([]string, map[string]interface{}, error) {
	if err := t.validate(); err != nil {
		return nil, nil, err
	}
	return tableFields, map[string]interface{}{
		logFieldMessage: t.message,
		logFieldHeaders: t.headers,
		logFieldData:    t.data,
	}, nil
}

func (t table) validate() error {
	if len(t.headers) == 0 {
		return errors.New("cannot create a table without headers")
	}
	return nil
}

func (t table) headerString() string {
	headers := make([]string, len(t.headers))
	for i, header := range t.headers {
		headers[i] = fmt.Sprintf("%s%s",
			color.New(color.Bold).SprintFunc()(header),
			strings.Repeat(" ", t.columnWidths[header]-len(header)),
		)
	}
	return Indent + strings.Join(headers, Gutter)
}

func (t table) dataString() string {
	rows := make([]string, len(t.data))
	for i, row := range t.data {
		cells := make([]string, len(t.headers))
		for j, header := range t.headers {
			cells[j] = fmt.Sprintf(
				"%s%s",
				row[header],
				strings.Repeat(" ", t.columnWidths[header]-len(row[header])),
			)
		}
		rows[i] = Indent + strings.Join(cells, Gutter)
	}
	return strings.Join(rows, "\n")
}

func (t table) dividerString() string {
	dashes := make([]string, len(t.headers))
	for i, header := range t.headers {
		dashes[i] = strings.Repeat("-", t.columnWidths[header])
	}
	return Indent + strings.Join(dashes, Gutter)
}

func parseValue(value interface{}) string {
	parsed := ""
	switch v := value.(type) {
	case nil: // leave zero-value
	case string:
		parsed = v
	case fmt.Stringer:
		parsed = v.String()
	case error:
		parsed = v.Error()
	default:
		parsed = fmt.Sprintf("%+v", v)
	}
	return parsed
}
