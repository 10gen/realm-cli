package cli

// DisableUsage disables the usage printing when an error occurs
type DisableUsage interface {
	DisableUsage() struct{}
}

type errDisableUsage struct {
	error
}

func (err errDisableUsage) DisableUsage() struct{} { return struct{}{} }

// CommandSuggester provides a list of suggested commands that will display to the user when an error occurs
type CommandSuggester interface {
	SuggestedCommands() []interface{}
}

// LinkReferrer provides a list of reference links that will display to the user when an error occurs
type LinkReferrer interface {
	ReferenceLinks() []interface{}
}
