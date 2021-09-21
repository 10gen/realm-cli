package terminal

import (
	"time"

	"github.com/briandowns/spinner"
)

// set of supported spinners
var (
	SpinnerCircles = []string{"㊂", "㊀", "㊁"}
	SpinnerDots    = []string{".  ", ".. ", "...", "   "}
)

// SpinnerOptions represents the spinner options
type SpinnerOptions struct {
	Icon     []string
	Duration time.Duration
}

// Spinner is a spinner
type Spinner interface {
	Start()
	Stop()
	SetMessage(message string)
}

type uiSpinner struct {
	s *spinner.Spinner
}

func newUISpinner(message string, opts SpinnerOptions) *uiSpinner {
	icon := opts.Icon
	if len(icon) == 0 {
		icon = SpinnerCircles
	}

	duration := opts.Duration
	if duration == 0 {
		duration = 250 * time.Millisecond
	}

	s := &uiSpinner{spinner.New(icon, duration)}
	s.SetMessage(message)
	return s
}

func (s *uiSpinner) Start()                    { s.s.Start() }
func (s *uiSpinner) Stop()                     { s.s.Stop() }
func (s *uiSpinner) SetMessage(message string) { s.s.Suffix = " " + message }

type noopSpinner struct{}

func (s noopSpinner) Start()                    {}
func (s noopSpinner) Stop()                     {}
func (s noopSpinner) SetMessage(message string) {}
