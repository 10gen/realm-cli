package ui

import (
	"os"
	"strings"

	"github.com/10gen/escaper"
	"github.com/mattn/go-isatty"
)

var ColorEnabled bool

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		ColorEnabled = true
	}
}

var esc = escaper.Default()

const (
	colorGroupStart = "%F{green}"
	colorGroupEnd   = "%f"

	colorAppClientIDStart = "%F{cyan}"
	colorAppClientIDEnd   = "%f"

	colorPermissionsStart = "%F{red}"
	colorPermissionsEnd   = "%f"

	colorServiceTypeStart = "%F{magenta}"
	colorServiceTypeEnd   = "%f"
)

type Variant int

const (
	None Variant = iota
	Group
	AppClientID
	Permissions
	ServiceType
)

func Color(v Variant, s string) string {
	if !ColorEnabled {
		return s
	}
	switch v {
	case Group:
		return group(s)
	case AppClientID:
		return appClientID(s)
	case Permissions:
		return permissions(s)
	case ServiceType:
		return serviceType(s)
	default:
		return s
	}
}

func group(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorGroupStart + s + colorGroupEnd)
}

func appClientID(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorAppClientIDStart + s + colorAppClientIDEnd)
}

func permissions(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorPermissionsStart + s + colorPermissionsEnd)
}

func serviceType(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorServiceTypeStart + s + colorServiceTypeEnd)
}
