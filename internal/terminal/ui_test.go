package terminal_test

import (
	"bytes"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestUIPrint(t *testing.T) {
	t.Run("Should select the correct Writer to Print with", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			log         terminal.Log
			expectedOut string
			expectedErr string
		}{
			{
				description: "Should use the default writer while printing an INFO log",
				log:         terminal.NewTextLog("test log"),
				expectedOut: "01:23:45 UTC INFO  test log\n",
			},
			{
				description: "Should use the error writer while printing an ERROR log",
				log:         terminal.NewErrorLog(errors.New("something bad happened")),
				expectedErr: "01:23:45 UTC ERROR something bad happened\n",
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				out, err := new(bytes.Buffer), new(bytes.Buffer)
				ui := terminal.NewUI(terminal.UIConfig{}, nil, out, err, logger)

				tc.log.Time = mock.StaticTime
				ui.Print(tc.log)

				assert.Equal(t, tc.expectedOut, out.String())
				assert.Equal(t, tc.expectedErr, err.String())
			})
		}
	})
}

var (
	logger = log.New(os.Stderr, "UTC ERROR ", log.Ltime|log.Lmsgprefix)
)
