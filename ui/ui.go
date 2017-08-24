// Package ui provides types and functions for conditionally coloring output,
// as well as for asking the command-line user questions.
package ui

import (
	"os"
	"strings"

	"github.com/10gen/escaper"
	"github.com/mattn/go-isatty"
)

// ColorEnabled determines whether or not any coloring (or other ANSI
// formatting) is done.
// By default, it gets set according to whether stdout is a TTY.
var ColorEnabled bool

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		ColorEnabled = true
	}
}

var esc = escaper.Default()

const (
	colorDiffOldStart = "%F{red}"
	colorDiffOldEnd   = "%f"

	colorDiffNewStart = "%F{green}"
	colorDiffNewEnd   = "%f"

	colorBooleanStart = "%F{red}"
	colorBooleanEnd   = "%f"

	colorGroupStart = "%F{green}"
	colorGroupEnd   = "%f"

	colorAppClientIDStart = "%F{cyan}"
	colorAppClientIDEnd   = "%f"

	colorPermissionsStart = "%F{red}"
	colorPermissionsEnd   = "%f"

	colorServiceTypeStart = "%F{magenta}"
	colorServiceTypeEnd   = "%f"

	colorParameterNameStart = "%B%F{yellow}"
	colorParameterNameEnd   = "%f%b"

	colorAuthProviderTypeStart = "%B%F{yellow}"
	colorAuthProviderTypeEnd   = "%f%b"
)

// Variant is a class of items for which colorings are defined.
type Variant int

// Variants that have been defined:
const (
	None Variant = iota
	DiffOld
	DiffNew
	Boolean
	Group
	AppClientID
	Permissions
	ServiceType
	ParameterName
	AuthProviderType
)

// Color applies a coloring corresponding to the supplied variant to the given
// string, provided that ColorEnabled is true.
func Color(v Variant, s string) string {
	if !ColorEnabled {
		return s
	}
	switch v {
	case DiffOld:
		return diffOld(s)
	case DiffNew:
		return diffNew(s)
	case Boolean:
		return boolean(s)
	case Group:
		return group(s)
	case AppClientID:
		return appClientID(s)
	case Permissions:
		return permissions(s)
	case ServiceType:
		return serviceType(s)
	case ParameterName:
		return parameterName(s)
	case AuthProviderType:
		return authProviderType(s)
	default:
		return s
	}
}

func diffOld(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorDiffOldStart + s + colorDiffOldEnd)
}

func diffNew(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorDiffNewStart + s + colorDiffNewEnd)
}

func boolean(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorBooleanStart + s + colorBooleanEnd)
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

func parameterName(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorParameterNameStart + s + colorParameterNameEnd)
}

func authProviderType(s string) string {
	s = strings.Replace(s, "%", "%%", -1)
	return esc.Expand(colorAuthProviderTypeStart + s + colorAuthProviderTypeEnd)
}
