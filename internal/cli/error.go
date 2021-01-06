package cli

// DisableUsage disables the usage printing when an error occurs
type DisableUsage interface {
	DisableUsage() struct{}
}

type errDisableUsage struct {
	error
}

func (err errDisableUsage) DisableUsage() struct{} { return struct{}{} }

// CommandSuggester handles any suggestions to run if the current command isn't working
type CommandSuggester interface {
	SuggestedCommands() []string
}

// LinkReferrer gives a list of links that relate to this command to give the user more context
type LinkReferrer interface {
	ReferenceLinks() []string
}
