package terminal

import (
	"fmt"
	"io"

	"github.com/10gen/realm-cli/internal/flags"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
)

// UIConfig holds the global config for the CLI ui
type UIConfig struct {
	DisableColors bool
	OutputFormat  string
	OutputTarget  string
}

// Messenger produces a message to display in the UI
type Messenger interface {
	Message() (string, error)
}

// UI is a terminal UI
type UI interface {
	AskOne(prompt survey.Prompt, answer interface{}) error
	Print(messages ...Messenger) error
}

type ui struct {
	config UIConfig
	err    io.Writer
	in     io.Reader
	out    io.Writer
}

// NewUI creates a new terminal UI
func NewUI(config UIConfig, in io.Reader, out, err io.Writer) UI {
	return &ui{
		config: config,
		err:    err,
		in:     in,
		out:    out,
	}
}

func (ui *ui) AskOne(prompt survey.Prompt, answer interface{}) error {
	stdio, stdioErr := ui.toStdio()
	if stdioErr != nil {
		return stdioErr
	}

	opts := survey.WithStdio(stdio.In, stdio.Out, stdio.Err)

	return survey.AskOne(prompt, answer, opts)
}

func (ui *ui) Print(messengers ...Messenger) error {
	color.NoColor = ui.config.DisableColors
	if ui.config.OutputFormat == flags.OutputFormatJSON {
		color.NoColor = true
	}

	for _, messenger := range messengers {
		message, err := messenger.Message()
		if err != nil {
			return err
		}
		fmt.Fprintln(ui.out, message)
	}
	return nil
}

func (ui *ui) toStdio() (terminal.Stdio, error) {
	in, inOK := ui.in.(terminal.FileReader)
	if !inOK {
		in = noopFdReader{ui.in}
	}
	out, outOK := ui.out.(terminal.FileWriter)
	if !outOK {
		out = noopFdWriter{ui.out}
	}
	return terminal.Stdio{
		In:  in,
		Out: out,
		Err: ui.err,
	}, nil
}

type noopFdReader struct {
	io.Reader
}

func (r noopFdReader) Fd() uintptr {
	return 0
}

type noopFdWriter struct {
	io.Writer
}

func (r noopFdWriter) Fd() uintptr {
	return 0
}
