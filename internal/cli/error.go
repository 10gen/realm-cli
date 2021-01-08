package cli

// DisableUsage disables the usage printing when an error occurs
type DisableUsage interface {
	DisableUsage() struct{}
}

type errDisableUsage struct {
	error
}

func (err errDisableUsage) DisableUsage() struct{} { return struct{}{} }

// CommandSuggester returns any suggested commands to remedy an error
type CommandSuggester interface {
	SuggestedCommands() []string
}

// LinkReferrer gives a list of reference links that are related to the error
type LinkReferrer interface {
	ReferenceLinks() []string
}
