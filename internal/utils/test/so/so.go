package so

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/smartystreets/goconvey/convey/gotest"
)

type failureView struct {
	Message  string `json:"Message"`
	Expected string `json:"Expected"`
	Actual   string `json:"Actual"`
}

var failedAssertionFormatter = color.New(color.FgYellow).SprintFunc()

// Assertion is a func that checks some condition for use in a test
type Assertion func(actual interface{}, expected ...interface{}) string

// So runs an assertion and fails the test if necessary
func So(t testing.TB, actual interface{}, assert Assertion, expected ...interface{}) {
	t.Helper()
	if result := assert(actual, expected...); result != "" {
		fv := failureView{}
		err := json.Unmarshal([]byte(result), &fv)
		errMessage := result
		if err == nil {
			errMessage = fv.Message
		}
		file, line, _ := gotest.ResolveExternalCaller()
		formatted := failedAssertionFormatter(fmt.Sprintf(
			"\nName: %s\n* %s\nLine %d:\n%s\n",
			t.Name(),
			file,
			line,
			errMessage,
		))

		t.Fatal(formatted)
	}
}
