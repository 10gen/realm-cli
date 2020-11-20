package mock

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/flags"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
)

// UIOptions are the options to configure the mock terminal UI
type UIOptions struct {
	UseColors bool
	UseJSON   bool
}

func newUIConfig(options UIOptions) terminal.UIConfig {
	outputFormat := flags.OutputFormatText
	if options.UseJSON {
		outputFormat = flags.OutputFormatJSON
	}

	return terminal.UIConfig{
		DisableColors: !options.UseColors,
		OutputFormat:  outputFormat,
	}
}

// NewUI creates a new mock terminal UI
func NewUI(options UIOptions, writer io.Writer) terminal.UI {
	return terminal.NewUI(
		newUIConfig(options),
		nil,
		writer,
		writer,
	)
}

// NewConsole creates a new *expect.Console along with its corresponding mock terminal UI
func NewConsole(options UIOptions, writers ...io.Writer) (*expect.Console, terminal.UI, error) {
	console, err := expect.NewConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, nil, err
	}

	ui := terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	)

	return console, ui, nil
}

// NewVT10XConsole creates a new *expect.Console along with its corresponding *vt10.State and xmock terminal UI
func NewVT10XConsole(options UIOptions, writers ...io.Writer) (*expect.Console, *vt10x.State, terminal.UI, error) {
	console, state, err := vt10x.NewVT10XConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, nil, nil, err
	}

	ui := terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	)

	return console, state, ui, nil
}

// FileWriter is a mock terminal UI writer which sends command output
// to a file named after the test being run
func FileWriter(t *testing.T) (*os.File, error) {
	// the ReplaceAll ensures no nested directories are needed when sub-tests are used
	filename := strings.ReplaceAll(fmt.Sprintf("%s.log", t.Name()), "/", "_")

	return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}
