package delete

import (
	"fmt"
	"strings"
)

const (
	flagUsers      = "user"
	flagUsersShort = "u"
	flagUsersUsage = "users to be deleted"

	flagInteractive      = "interactive-filter"
	flagInteractiveShort = "i"
	flagInteractiveUsage = "filter users interactively"

	flagStatusType      = "pending"
	flagStatusTypeShort = "p"
	flagStatusTypeUsage = `select the users' status: ["confirmed", "pending"]`
)

type statusType string

// String returns the status type display
func (st statusType) String() string { return string(st) }

// Type returns the statusType type
func (st statusType) Type() string { return "string" }

// Set validates and sets the status type value
func (st *statusType) Set(val string) error {
	newStatusType := statusType(val)

	if !isValidStatusType(newStatusType) {
		return errInvalidStatusType
	}

	*st = newStatusType
	return nil
}

const (
	statusTypeNil       statusType = ""
	statusTypeConfirmed statusType = "confirmed"
	statusTypePending   statusType = "pending"
)

var (
	errInvalidStatusType = func() error {
		allStatusTypes := []string{statusTypeConfirmed.String(), statusTypePending.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allStatusTypes, ", "))
	}()
)

func isValidStatusType(st statusType) bool {
	switch st {
	case
		statusTypeNil, // allow statusType to be optional
		statusTypeConfirmed,
		statusTypePending:
		return true
	}
	return false
}
