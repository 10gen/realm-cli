package config

import "errors"

var (
	ErrAlreadyLoggedIn = errors.New("You are already logged in.")
)

// LoggedIn checks whether the local config has a logged in user.
func LoggedIn() bool {
	return false // TODO
}

func LogIn(token string) error {
	if LoggedIn() {
		return ErrAlreadyLoggedIn
	}
	return nil // TODO
}
