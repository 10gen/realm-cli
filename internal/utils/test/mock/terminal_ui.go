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

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
)

// UIOptions are the options to configure the mock terminal UI
type UIOptions struct {
	AutoConfirm   bool
	UseColors     bool
	UseJSON       bool
	OpenBrowserFn func(url string) error
}

func newUIConfig(options UIOptions) terminal.UIConfig {
	outputFormat := terminal.OutputFormatText
	if options.UseJSON {
		outputFormat = terminal.OutputFormatJSON
	}

	return terminal.UIConfig{
		AutoConfirm:   options.AutoConfirm,
		DisableColors: !options.UseColors,
		OutputFormat:  outputFormat,
	}
}

var (
	// StaticTime represents a time.Time that displays the clock as 01:23:45.678
	StaticTime = time.Date(1989, 6, 22, 1, 23, 45, 0, time.UTC)
)

type ui struct {
	terminal.UI
	OpenBrowserFn func(url string) error
}

func (ui ui) OpenBrowser(url string) error {
	if ui.OpenBrowserFn == nil {
		return ui.UI.OpenBrowser(url)
	}
	return ui.OpenBrowserFn(url)
}

func (ui ui) Print(logs ...terminal.Log) {
	for i := range logs {
		logs[i].Time = StaticTime
	}
	ui.UI.Print(logs...)
}

// NewUI returns a new *bytes.Buffer and a mock terminal UI that writes to the buffer
func NewUI() (*bytes.Buffer, terminal.UI) {
	out := new(bytes.Buffer)
	return out, NewUIWithOptions(UIOptions{}, out)
}

// NewUIWithOptions creates a new mock terminal UI based on the provided options
func NewUIWithOptions(options UIOptions, writer io.Writer) terminal.UI {
	return ui{terminal.NewUI(
		newUIConfig(options),
		nil,
		writer,
		writer,
	), options.OpenBrowserFn}
}

// NewConsole returns a new *bytes.Buffer and a *expect.Console
// along with its corresponding mock terminal UI that write to the buffer
func NewConsole() (*bytes.Buffer, *expect.Console, terminal.UI, error) {
	out := new(bytes.Buffer)
	console, ui, err := NewConsoleWithOptions(UIOptions{}, out)
	return out, console, ui, err
}

// NewConsoleWithOptions creates a new *expect.Console
// along with its corresponding mock terminal UI based on the provided options
func NewConsoleWithOptions(options UIOptions, writers ...io.Writer) (*expect.Console, terminal.UI, error) {
	console, err := expect.NewConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, nil, err
	}

	ui := ui{terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	), options.OpenBrowserFn}

	return console, ui, nil
}

// NewVT10XConsole returns a new *bytes.Buffer and a *expect.Console
// along with its corresponding *vt10.State and mock terminal UI that write to the buffer
func NewVT10XConsole() (*bytes.Buffer, *expect.Console, *vt10x.State, terminal.UI, error) {
	out := new(bytes.Buffer)
	console, state, ui, err := NewVT10XConsoleWithOptions(UIOptions{}, out)
	return out, console, state, ui, err
}

// NewVT10XConsoleWithOptions creates a new *expect.Console
// along with its corresponding *vt10.State and mock terminal UI based on the provided options
func NewVT10XConsoleWithOptions(options UIOptions, writers ...io.Writer) (*expect.Console, *vt10x.State, terminal.UI, error) {
	console, state, err := vt10x.NewVT10XConsole(expect.WithStdout(writers...))
	if err != nil {
		return nil, nil, nil, err
	}

	ui := ui{terminal.NewUI(
		newUIConfig(options),
		console.Tty(),
		console.Tty(),
		console.Tty(),
	), options.OpenBrowserFn}

	return console, state, ui, nil
}

// FileWriter is a mock terminal UI writer which sends command output
// to a file named after the test being run
func FileWriter(t *testing.T) (*os.File, error) {
	// the ReplaceAll ensures no nested directories are needed when sub-tests are used
	filename := strings.ReplaceAll(fmt.Sprintf("%s.log", t.Name()), "/", "_")

	return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}
