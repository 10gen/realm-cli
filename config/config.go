package config

import "errors"

var (
	ErrAlreadyLoggedIn = errors.New("stitch: you are already logged in.")
	ErrNotLoggedIn     = errors.New("stitch: you are not logged in.")
)

var Chdir string

// LoggedIn checks whether the local config has a logged in user.
func LoggedIn() bool {
	return true // TODO
}

func LogIn(token string) error {
	if LoggedIn() {
		return ErrAlreadyLoggedIn
	}
	return nil // TODO
}

func LogOut() error {
	if !LoggedIn() {
		return ErrNotLoggedIn
	}
	return nil // TODO
}
