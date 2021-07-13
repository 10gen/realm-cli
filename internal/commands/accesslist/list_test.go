package accesslist

import (
	"errors"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestAllowedIPListHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}
	testAllowedIPs := realm.AccessList{
		[]realm.AllowedIP{
			{Address: "0.0.0.0", Comment: "comment"},
			{Address: "192.1.1.1"},
			{Address: "192.158.1.38", Comment: "cool comment"},
		},
		"0.0.0.0",
	}

	for _, tc := range []struct {
		description    string
		accessList     realm.AccessList
		expectedOutput string
	}{
		{
			description:    "should list no allowed IPs with no app allowed IPs found",
			expectedOutput: "No available allowed IPs to show\n",
		},
		{
			description: "should list the allowed IPs found for the app",
			accessList:  testAllowedIPs,
			expectedOutput: strings.Join(
				[]string{
					"Found 3 allowed IPs",
					"  IP Address    Comment     ",
					"  ------------  ------------",
					"  0.0.0.0       comment     ",
					"  192.1.1.1                 ",
					"  192.158.1.38  cool comment",
					"",
				},
				"\n",
			),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}

			realmClient.AllowedIPsFn = func(groupID, appID string) (realm.AccessList, error) {
				return tc.accessList, nil
			}

			cmd := &CommandList{listInputs{cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			}}}

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "when resolving the app fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
			{
				description: "when finding the allowed IPs fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.AllowedIPsFn = func(groupID, appID string) (realm.AccessList, error) {
						return realm.AccessList{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandList{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
