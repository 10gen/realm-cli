package terminal

import (
	"errors"
	"fmt"
	"strings"
)

type followUp struct {
	stringRep string
	name      string
}

func (f followUp) String() string {
	return f.stringRep
}

type FollowUpType struct {
	LINK    *followUp
	COMMAND *followUp

	types []*followUp
}

const (
	FollowUpLink    = "FollowUpLink"
	FollowUpCommand = "FollowUpCommand"
)

const (
	linkMessage = "For more information"
	cmdMessage  = "Try running instead"

	// sep is the separation between links or suggested commands
	sep = ", "
)

var (
	followUpFields = []string{logFieldMessage}
)

var followUpTypes = newFollowUpType()

func newFollowUpType() *FollowUpType {
	link := &followUp{FollowUpLink, linkMessage}
	cmd := &followUp{FollowUpCommand, cmdMessage}
	return &FollowUpType{
		LINK:    link,
		COMMAND: cmd,

		types: []*followUp{link, cmd},
	}
}

func (ft FollowUpType) parse(key string) (*followUp, error) {
	for _, t := range ft.types {
		if t.String() == key {
			return t, nil
		}
	}
	return nil, errors.New("cannot find type in FollowUpTypes: " + key)
}

//
func (ft FollowUpType) contains(key string) bool {
	for _, t := range ft.types {
		if t.String() == key {
			return true
		}
	}
	return false
}

type FollowUpMessage struct {
	messageType *followUp
	message     string
}

func (fm FollowUpMessage) validate() error {
	if fm.messageType == nil || len(fm.message) == 0 {
		return errors.New("empty follow up message")
	}
	if _, err := followUpTypes.parse(fm.messageType.name); err != nil {
		return err
	}
	return nil
}

func NewFollowUpMessage(followUpType string, messages []string) FollowUpMessage {
	var f FollowUpMessage
	parsed, err := followUpTypes.parse(followUpType)
	if err != nil {
		return f
	}

	f.message = fmt.Sprintf(`%s: %s`, parsed.name, strings.Join(messages, sep))
	f.messageType = parsed

	return f
}

func (fm FollowUpMessage) Message() (string, error) {
	if err := fm.validate(); err != nil {
		return "", err
	}
	return fm.message, nil
}

func (fm FollowUpMessage) Payload() ([]string, map[string]interface{}, error) {
	if err := fm.validate(); err != nil {
		return nil, nil, err
	}
	return followUpFields, map[string]interface{}{
		logFieldMessage: fm.message,
	}, nil
}
