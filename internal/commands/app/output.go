package app

import "github.com/10gen/realm-cli/internal/cloud/realm"

type appOutput struct {
	app realm.App
	err error
}

const (
	headerID      = "ID"
	headerName    = "Name"
	headerDeleted = "Deleted"
	headerDetails = "Details"
)
