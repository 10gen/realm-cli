package feedback

import (
	"fmt"
)

// ErrUsageHider specifies if the command's usage should be hidden when an error occurs
type ErrUsageHider interface {
	HideUsage() bool
}

// ErrSuggester provides a list of suggestions that will display to the user when an error occurs
type ErrSuggester interface {
	Suggestions() []interface{}
}

// ErrLinkReferrer provides a list of reference links that will display to the user when an error occurs
type ErrLinkReferrer interface {
	ReferenceLinks() []interface{}
}

// NewErr returns a new CLI error
func NewErr(cause error, details ...ErrDetail) error {
	var d ErrDetails
	for _, detail := range details {
		detail.ApplyTo(&d)
	}

	var hideUsage bool
	if d.HideUsage != nil {
		hideUsage = *d.HideUsage
	}

	return cliErr{
		cause:          cause,
		hideUsage:      hideUsage,
		suggestions:    d.Suggestions,
		referenceLinks: d.ReferenceLinks,
	}
}

// WrapErr wraps the provided error with the provided message
// The extra details are applied on top of any existing, wrapped details
// Note: it is assumed the message contains a single `fmt` verb to include the cause/error
func WrapErr(msgFmt string, cause error, details ...ErrDetail) error {
	err, ok := cause.(cliErr)
	if !ok {
		return NewErr(fmt.Errorf(msgFmt, cause), details...)
	}

	var d ErrDetails
	for _, detail := range details {
		detail.ApplyTo(&d)
	}

	if d.HideUsage == nil {
		d.HideUsage = &err.hideUsage
	}
	d.Suggestions = append(d.Suggestions, err.suggestions...)
	d.ReferenceLinks = append(d.ReferenceLinks, err.referenceLinks...)

	newDetails := make([]ErrDetail, 0, 1+len(d.Suggestions)+len(d.ReferenceLinks))
	if d.HideUsage != nil {
		newDetails = append(newDetails, ErrNoUsage{})
	}
	for _, suggestion := range d.Suggestions {
		newDetails = append(newDetails, ErrSuggestion{suggestion})
	}
	for _, link := range d.ReferenceLinks {
		newDetails = append(newDetails, ErrReferenceLink{link})
	}

	return NewErr(fmt.Errorf(msgFmt, cause), newDetails...)
}

type cliErr struct {
	cause          error
	hideUsage      bool
	suggestions    []interface{}
	referenceLinks []interface{}
}

func (err cliErr) Error() string {
	return err.cause.Error()
}

func (err cliErr) HideUsage() bool {
	return err.hideUsage
}

func (err cliErr) Suggestions() []interface{} {
	return err.suggestions
}

func (err cliErr) ReferenceLinks() []interface{} {
	return err.referenceLinks
}

// ErrDetails represent a CLI error's details
type ErrDetails struct {
	HideUsage      *bool
	Suggestions    []interface{}
	ReferenceLinks []interface{}
}

// ErrDetail represents a single detail of a CLI error
type ErrDetail interface {
	ApplyTo(details *ErrDetails)
}

// ErrNoUsage hides the command's usage for an error
type ErrNoUsage struct{}

// ApplyTo will set error to hide usage
func (err ErrNoUsage) ApplyTo(details *ErrDetails) {
	t := true
	details.HideUsage = &t
}

// ErrSuggestion adds a suggestion to an error
type ErrSuggestion struct {
	Suggestion interface{}
}

// ApplyTo will add the suggestion to the error details
func (err ErrSuggestion) ApplyTo(details *ErrDetails) {
	details.Suggestions = append(details.Suggestions, err.Suggestion)
}

// ErrReferenceLink adds a reference link to an error
type ErrReferenceLink struct {
	Link interface{}
}

// ApplyTo will add the reference link to the error details
func (err ErrReferenceLink) ApplyTo(details *ErrDetails) {
	details.ReferenceLinks = append(details.ReferenceLinks, err.Link)
}
