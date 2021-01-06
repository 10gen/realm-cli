package terminal

import (
	"reflect"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestNewFollowUpMessage(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(followUpMessage{}), cmp.AllowUnexported(followUpMessage{}))

	for _, tc := range []struct {
		description      string
		message          string
		followUps        []string
		expectedFollowUp followUpMessage
	}{
		{
			description: "Should return a follow up message even if there is no message",
			message:     "",
			followUps:   []string{"existing follow up"},
			expectedFollowUp: followUpMessage{
				"",
				[]string{"existing follow up"},
			},
		},
		{
			description: "Should return a follow up message even if there are no followups",
			message:     CommandMessage,
			followUps:   []string{},
			expectedFollowUp: followUpMessage{
				CommandMessage,
				[]string{},
			},
		},
		{
			description: "Should return a follow up message",
			message:     LinkMessage,
			followUps:   []string{"follow", "up"},
			expectedFollowUp: followUpMessage{
				LinkMessage,
				[]string{"follow", "up"},
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedFollowUp, NewFollowUpMessage(tc.message, tc.followUps))
		})
	}
}

func TestFollowUpMessage(t *testing.T) {

	for _, tc := range []struct {
		description     string
		followUpMessage followUpMessage
	}{
		{
			description:     "Should return an error if there is no message",
			followUpMessage: followUpMessage{"", []string{"something"}},
		},
		{
			description:     "Should return an error if there are no followUps",
			followUpMessage: followUpMessage{LinkMessage, nil},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := tc.followUpMessage.Message()
			assert.Equal(t, "", actual)
			assert.Equal(t, err.Error(), "empty follow up message")
		})
	}

	for _, tc := range []struct {
		description     string
		followUpMessage followUpMessage
		expectedMessage string
	}{
		{
			description:     "Should print a message for one followUp on the same line",
			followUpMessage: followUpMessage{LinkMessage, []string{"https://mongodb.com"}},
			expectedMessage: `Refer to the following link https://mongodb.com`,
		},
		{
			description:     "Should print a message for multiple followUps on multiple lines with plurals",
			followUpMessage: followUpMessage{CommandMessage, []string{"cwd", "ls", "mkdir"}},
			expectedMessage: strings.Join(
				[]string{
					"Try the following commands",
					"  cwd",
					"  ls",
					"  mkdir",
				}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actualMessage, err := tc.followUpMessage.Message()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedMessage, actualMessage)
		})
	}
}

func TestFollowUpPayload(t *testing.T) {
	for _, tc := range []struct {
		description     string
		followUpMessage followUpMessage
	}{
		{
			description:     "Should return an error if there is no message",
			followUpMessage: followUpMessage{"", []string{"something"}},
		},
		{
			description:     "Should return an error if there is no followUp",
			followUpMessage: followUpMessage{LinkMessage, nil},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actualFields, actualPayload, err := tc.followUpMessage.Payload()
			assert.Nil(t, actualFields)
			assert.Nil(t, actualPayload)
			assert.Equal(t, err.Error(), "empty follow up message")
		})
	}

	t.Run("Should return a payload for a valid followUpMessage", func(t *testing.T) {
		expectedFollowUps := []string{"https://mongodb.com"}

		followUp := followUpMessage{
			LinkMessage,
			expectedFollowUps,
		}
		actualFields, actualPayload, err := followUp.Payload()
		assert.Nil(t, err)
		assert.Equal(t, followUpFields, actualFields)

		assert.Equal(t, LinkMessage, actualPayload[logFieldMessage])
		assert.Equal(t, expectedFollowUps, actualPayload[logFieldFollowUps])
	})
}
