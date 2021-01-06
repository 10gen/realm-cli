package terminal

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CommandMessage = "Try the following command"
	LinkMessage    = "Refer to the following link"

	logFieldFollowUps = "followUps"
)

var (
	followUpFields = []string{logFieldMessage, logFieldFollowUps}
)

type followUpMessage struct {
	message   string
	followUps []string
}

func (fm followUpMessage) validate() error {
	if len(fm.message) == 0 || len(fm.followUps) == 0 {
		return errors.New("empty follow up message")
	}
	return nil
}

func NewFollowUpMessage(message string, followUps []string) followUpMessage {
	return followUpMessage{
		message:   message,
		followUps: followUps,
	}
}

func (fm followUpMessage) Message() (string, error) {
	if err := fm.validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", fm.formatMessage(), fm.formatFollowUp()), nil
}

func (fm followUpMessage) Payload() ([]string, map[string]interface{}, error) {
	if err := fm.validate(); err != nil {
		return nil, nil, err
	}
	return followUpFields, map[string]interface{}{
		logFieldMessage:   fm.message,
		logFieldFollowUps: fm.followUps,
	}, nil
}

func (fm followUpMessage) formatMessage() string {
	if len(fm.followUps) > 1 {
		return fm.message + "s"
	}
	return fm.message
}

// TODO: similar to the dataString function in list; consolidate?
func (fm followUpMessage) formatFollowUp() string {
	if len(fm.followUps) == 1 {
		return " " + fm.followUps[0]
	}
	followUps := make([]string, 0, len(fm.followUps)+1)
	followUps = append(followUps, "")
	for _, f := range fm.followUps {
		followUps = append(followUps, indent+f)
	}
	return strings.Join(followUps, "\n")
}
