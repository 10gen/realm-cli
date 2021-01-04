package cli

import (
	"github.com/10gen/realm-cli/internal/terminal"
)

type followUp struct {
	followUpType terminal.FollowUp
	messages []string
}

//type CommandSuggester interface {
//	SuggestedCommands() []string
//}
//
//type LinkReferrer interface {
//	ReferenceLinks() []string
//}

func (f followUp) SuggestedCommands() []string{
	return f.messages
}

func (f followUp) ReferenceLinks() []string {
	return f.messages
}