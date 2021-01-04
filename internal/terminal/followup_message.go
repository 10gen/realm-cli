package terminal

type FollowUp struct {
	message string
}

type FollowUpType struct {
	LINK FollowUp
	COMMAND FollowUp

	followUpTypes []FollowUp
}

const (
	linkMessage = "For more information"
	cmdMessage = "Try running instead"
)

func newFollowUpType() *FollowUpType{
	link := FollowUp{linkMessage}
	cmd := FollowUp{cmdMessage}
	return &FollowUpType{
		LINK: link,
		COMMAND: cmd,
		followUpTypes: []FollowUp{link, cmd},
	}
}

var FollowUpTypes = newFollowUpType()

type followUpMessage struct {
	messageType FollowUp
	message string
}

func newFollowUpMessage(followUp FollowUp, messages []string) followUpMessage {

}

func validate() error {
	// TODO: if there is a proper followup type, if there are even links / suggested commands to run
}

func (fm followUpMessage) Message() (string, error) {

}

func (fm followUpMessage) Payload() ([]string, map[string]interface{}, error) {

}