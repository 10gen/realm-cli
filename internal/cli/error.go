package cli

// DisableUsage disables the usage printing when an error occurs
type DisableUsage interface {
	DisableUsage() struct{}
}

type errDisableUsage struct {
	error
}

func (err errDisableUsage) DisableUsage() struct{} { return struct{}{} }

// Suggester provides a list of suggestions that will display to the user when an error occurs
type Suggester interface {
	Suggestions() []interface{}
}

// LinkReferrer provides a list of reference links that will display to the user when an error occurs
type LinkReferrer interface {
	ReferenceLinks() []interface{}
}
