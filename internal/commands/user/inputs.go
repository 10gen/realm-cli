package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/AlecAivazis/survey/v2"
)

const (
	userDataEmail = "email"
	userDataName  = "name"
)

type multiUserInputs struct {
	State         realm.UserState
	ProviderTypes []string
	Pending       bool
	Users         []string
}

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

func (i multiUserInputs) filter() realm.UserFilter {
	return realm.UserFilter{
		IDs:       i.Users,
		State:     i.State,
		Pending:   i.Pending,
		Providers: realm.NewAuthProviderTypes(i.ProviderTypes...),
	}
}

func (i multiUserInputs) resolveUsers(realmClient realm.Client, groupID, appID string) ([]realm.User, error) {
	foundUsers, findErr := realmClient.FindUsers(groupID, appID, i.filter())
	if findErr != nil {
		return nil, findErr
	}
	if len(i.Users) > 0 && len(foundUsers) == 0 {
		return nil, errors.New("no users found")
	}
	return foundUsers, nil
}

func (i multiUserInputs) selectUsers(ui terminal.UI, resolvedUsers []realm.User, action string) ([]realm.User, error) {
	if len(i.Users) > 0 {
		return resolvedUsers, nil
	}
	selectableUsers := map[string]realm.User{}
	selectableUserOptions := make([]string, len(resolvedUsers))
	for idx, user := range resolvedUsers {
		var apt realm.AuthProviderType
		if len(user.Identities) > 0 {
			apt = user.Identities[0].ProviderType
		}
		opt := displayUser(apt, user)
		selectableUserOptions[idx] = opt
		selectableUsers[opt] = user
	}
	var selectedUsers []string
	askErr := ui.AskOne(
		&selectedUsers,
		&survey.MultiSelect{
			Message: fmt.Sprintf("Which user(s) would you like to %s?", action),
			Options: selectableUserOptions,
		},
	)
	if askErr != nil {
		return nil, askErr
	}
	users := make([]realm.User, len(selectedUsers))
	for idx, user := range selectedUsers {
		users[idx] = selectableUsers[user]
	}
	return users, nil
}
