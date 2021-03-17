package atlas

import (
	"errors"
	"fmt"
)

var (
	errCommonServerError = "an unexpected server error has occurred"

	errCommonUnauthorized = "failed to authenticate with MongoDB Cloud API"

	errCommonForbidden = "Please check your Atlas API Whitelist entries to " +
		"ensure that requests from this IP address are allowed"
)

// set of known MongoDB Cloud Atlas errors
var (
	ErrMissingAuth = errors.New("must provide auth details")
)

type errResponse struct {
	Detail    string `json:"detail"`
	Error     int    `json:"error"`
	ErrorCode string `json:"errorCode"`
}

type errServerError struct {
	reason string
}

func (err errServerError) Error() string {
	if err.reason == "" {
		return errCommonServerError
	}
	return fmt.Sprintf("%s: %s", errCommonServerError, err.reason)
}

// ErrUnauthorized is an unauthorized error
type ErrUnauthorized struct {
	Reason string
}

func (err ErrUnauthorized) Error() string {
	return fmt.Sprintf("%s: %s", errCommonUnauthorized, err.Reason)
}

type errForbidden struct {
	status string
}

func (err errForbidden) Error() string {
	return fmt.Sprintf("(%s) %s", err.status, errCommonForbidden)
}

func (err errForbidden) ReferenceLinks() []interface{} {
	return []interface{}{
		"https://cloud.mongodb.com/v2#/account/publicApi",
	}
}
