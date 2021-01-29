package user

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	userDataEmail = "email"
	userDataName  = "name"
)

func validAuthProviderTypes() []interface{} {
	apts := make([]interface{}, 0, len(realm.ValidAuthProviderTypes))
	for _, apt := range realm.ValidAuthProviderTypes {
		apts = append(apts, apt)
	}
	return apts
}

func displayUser(apt realm.AuthProviderType, user realm.User) string {
	var sb strings.Builder
	sb.WriteString(apt.Display() + terminal.DelimiterInline)
	if data := displayUserData(apt, user); data != "" {
		sb.WriteString(data + terminal.DelimiterInline)
	}
	sb.WriteString(user.ID)
	return sb.String()
}

func displayUserData(apt realm.AuthProviderType, user realm.User) string {
	var (
		val interface{}
		ok  bool
	)
	switch apt {
	case realm.AuthProviderTypeUserPassword:
		val, ok = user.Data["email"]
	case realm.AuthProviderTypeAPIKey:
		val, ok = user.Data["name"]
	}
	if ok {
		return fmt.Sprint(val)
	}
	return ""
}
