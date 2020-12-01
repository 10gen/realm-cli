package terminal

import (
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
)

// UI is a terminal UI
type UI interface {
	AskOne(prompt survey.Prompt, answer interface{}) error
	Print(logs ...Log) error
}

// NewUI creates a new terminal UI
func NewUI(config UIConfig, in io.Reader, out, err io.Writer) UI {
	noColor := config.DisableColors
	if config.OutputFormat == OutputFormatJSON {
		noColor = true
	}
	color.NoColor = noColor

	return &ui{
		config: config,
		err:    err,
		in:     in,
		out:    out,
	}
}

type ui struct {
	config UIConfig
	err    io.Writer
	in     io.Reader
	out    io.Writer
}

// UIConfig holds the global config for the CLI ui
type UIConfig struct {
	DisableColors bool
	OutputFormat  OutputFormat
	OutputTarget  string
}

func (ui *ui) AskOne(prompt survey.Prompt, answer interface{}) error {
	stdio, stdioErr := ui.toStdio()
	if stdioErr != nil {
		return stdioErr
	}

	opts := survey.WithStdio(stdio.In, stdio.Out, stdio.Err)
	return survey.AskOne(prompt, answer, opts)
}

func (ui *ui) Print(logs ...Log) error {
	for _, log := range logs {
		output, outputErr := log.Print(ui.config.OutputFormat)
		if outputErr != nil {
			return outputErr
		}

		var writer io.Writer
		switch log.Level {
		case LogLevelError:
			writer = ui.err
		default:
			writer = ui.out
		}

		if _, err := fmt.Fprintln(writer, output); err != nil {
			return err
		}
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
