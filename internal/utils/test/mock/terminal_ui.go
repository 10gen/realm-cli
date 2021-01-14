package mock

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
)

// UIOptions are the options to configure the mock terminal UI
type UIOptions struct {
	UseColors bool
	UseJSON   bool
}

func newUIConfig(options UIOptions) terminal.UIConfig {
	outputFormat := terminal.OutputFormatText
	if options.UseJSON {
		outputFormat = terminal.OutputFormatJSON
	}

	return terminal.UIConfig{
		DisableColors: !options.UseColors,
		OutputFormat:  outputFormat,
	}
}

var (
	// StaticTime represents a time.Time that displays the clock as 01:23:45.678
	StaticTime = time.Date(1989, 6, 22, 1, 23, 45, 0, time.UTC)
)

// UI is a mocked terminal.UI
type UI struct {
	terminal.UI

	AskOneFn func(answer interface{}, prompt survey.Prompt) error
}

// Print sets the time and then calls the terminal.UI.Print
func (ui UI) Print(logs ...terminal.Log) error {
	for i := range logs {
		logs[i].Time = StaticTime
	}
	return ui.UI.Print(logs...)
}

// NewUI returns a new *bytes.Buffer and a mock terminal UI that writes to the buffer
func NewUI() (*bytes.Buffer, UI) {
	out := new(bytes.Buffer)
	return out, NewUIWithOptions(UIOptions{}, out)
}

// NewUIWithOptions creates a new mock terminal UI based on the provided options
func NewUIWithOptions(options UIOptions, writer io.Writer) UI {
	return UI{terminal.NewUI(
		newUIConfig(options),
		nil,
		writer,
		writer,
	), nil}
}

// NewConsole returns a new *bytes.Buffer and a *expect.Console
// along with its corresponding mock terminal UI that write to the buffer
func NewConsole() (*bytes.Buffer, *expect.Console, UI, error) {
	out := new(bytes.Buffer)
	console, ui, err := NewConsoleWithOptions(UIOptions{}, out)
	return out, console, ui, err
}

// NewConsoleWithOptions creates a new *expect.Console
// along with its corresponding mock terminal UI based on the provided options
func NewConsoleWithOptions(options UIOptions, writers ...io.Writer) (*expect.Console, UI, error) {
	console, err := expect.NewConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, UI{}, err
	}

	ui := UI{terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	), nil}

	return console, ui, nil
}

// NewVT10XConsole returns a new *bytes.Buffer and a *expect.Console
// along with its corresponding *vt10.State and mock terminal UI that write to the buffer
func NewVT10XConsole() (*bytes.Buffer, *expect.Console, *vt10x.State, UI, error) {
	out := new(bytes.Buffer)
	console, state, ui, err := NewVT10XConsoleWithOptions(UIOptions{}, out)
	return out, console, state, ui, err
}

// NewVT10XConsoleWithOptions creates a new *expect.Console
// along with its corresponding *vt10.State and mock terminal UI based on the provided options
func NewVT10XConsoleWithOptions(options UIOptions, writers ...io.Writer) (*expect.Console, *vt10x.State, UI, error) {
	console, state, err := vt10x.NewVT10XConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, nil, UI{}, err
	}

	ui := UI{terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	), nil}

	return console, state, ui, nil
}

// FileWriter is a mock terminal UI writer which sends command output
// to a file named after the test being run
func FileWriter(t *testing.T) (*os.File, error) {
	// the ReplaceAll ensures no nested directories are needed when sub-tests are used
	filename := strings.ReplaceAll(fmt.Sprintf("%s.log", t.Name()), "/", "_")

	return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

// AskOne calls the mocked AskOne implementation if provided,
// otherwise the call falls back to the underlying terminal.UI implementation.
// NOTE: this may panic if the underlying terminal.UI is left undefined
func (ui UI) AskOne(answer interface{}, prompt survey.Prompt) error {
	if ui.AskOneFn != nil {
		return ui.AskOneFn(answer, prompt)
	}
	return ui.UI.AskOne(answer, prompt)
}
