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

func TestAllowedIPDeleteHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "test-abcde",
		Name:        "test",
	}

	allowedIPs := []realm.AllowedIP{
		{ID: "address1", Address: "0.0.0.0", Comment: "comment"},
		{ID: "address2", Address: "192.1.1.1"},
		{ID: "address3", Address: "192.158.1.38", Comment: "cool comment"},
	}

	t.Run("should show empty state message if no allowed ips are found", func(t *testing.T) {
		_, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		realmClient.AllowedIPsFn = func(groupID, appID string) ([]realm.AllowedIP, error) {
			return nil, nil
		}

		cmd := &CommandDelete{deleteInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
		}}

		assert.Equal(t, errors.New("no IP addresses or CIDR blocks to delete"), cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
	})

	for _, tc := range []struct {
		description          string
		testInput            []string
		expectedOutput       string
		deleteErr            error
		expectedAllowedIPIDs []string
	}{
		{
			description: "should return successful outputs for proper inputs",
			testInput:   []string{"0.0.0.0", "192.1.1.1"},
			expectedOutput: strings.Join([]string{
				"Deleted 2 IP address(es) and/or CIDR block(s)",
				"  IP Address  Comment  Deleted  Details",
				"  ----------  -------  -------  -------",
				"  0.0.0.0     comment  true            ",
				"  192.1.1.1            true            ",
				"",
			}, "\n"),
			expectedAllowedIPIDs: []string{"address1", "address2"},
		},
		{
			description: "should output the errors for deletes on individual allowed ips",
			testInput:   []string{"0.0.0.0", "192.158.1.38"},
			expectedOutput: strings.Join([]string{
				"Deleted 2 IP address(es) and/or CIDR block(s)",
				"  IP Address    Comment       Deleted  Details           ",
				"  ------------  ------------  -------  ------------------",
				"  0.0.0.0       comment       false    something happened",
				"  192.158.1.38  cool comment  false    something happened",
				"",
			}, "\n"),
			deleteErr:            errors.New("something happened"),
			expectedAllowedIPIDs: []string{"address1", "address3"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			var deleteArgs struct {
				groupID, appID, allowedIPID string
			}

			var allowedIPIDs []string

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}
			realmClient.AllowedIPsFn = func(groupID, appID string) ([]realm.AllowedIP, error) {
				return allowedIPs, nil
			}
			realmClient.AllowedIPDeleteFn = func(groupID, appID, allowedIPID string) error {
				deleteArgs = struct {
					groupID, appID, allowedIPID string
				}{groupID, appID, allowedIPID}
				allowedIPIDs = append(allowedIPIDs, allowedIPID)
				return tc.deleteErr
			}

			cmd := &CommandDelete{deleteInputs{
				cli.ProjectInputs{Project: projectID, App: appID},
				tc.testInput,
			}}

			out, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, tc.expectedOutput, out.String())
			assert.Equal(t, "projectID", deleteArgs.groupID)
			assert.Equal(t, "appID", deleteArgs.appID)
			assert.Equal(t, tc.expectedAllowedIPIDs, allowedIPIDs)
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			realmClient func() realm.Client
			testInput   []string
			expectedErr error
		}{
			{
				description: "if there is an issue with finding allowed ips",
				realmClient: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return nil, errors.New("something happened with allowed ips")
						},
					}
				},
				expectedErr: errors.New("something happened with allowed ips"),
			},
			{
				description: "if there is no app",
				realmClient: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return nil, errors.New("something went wrong with the app")
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return allowedIPs, nil
						},
					}
				},
				expectedErr: errors.New("something went wrong with the app"),
			},
			{
				description: "if provided ip address or cidr block does not exist in access list",
				realmClient: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return allowedIPs, nil
						},
					}
				},
				testInput:   []string{"0.0.0.0", "0.0.0.1", "192.1.2.1"},
				expectedErr: errors.New("unable to find IP address(es) and/or CIDR block(s): 0.0.0.1, 192.1.2.1"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				cmd := &CommandDelete{deleteInputs{
					cli.ProjectInputs{Project: projectID, App: appID},
					tc.testInput,
				}}
				err := cmd.Handler(nil, nil, cli.Clients{Realm: tc.realmClient()})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}

func TestAllowedIPDeleteInputs(t *testing.T) {
	allowedIPs := []realm.AllowedIP{
		{ID: "address1", Address: "0.0.0.0", Comment: "comment"},
		{ID: "address2", Address: "192.1.1.1"},
		{ID: "address3", Address: "192.158.1.38", Comment: "cool comment"},
	}

	t.Run("should return allowed ips when specified by address", func(t *testing.T) {
		inputs := deleteInputs{
			Addresses: []string{"0.0.0.0"},
		}

		allowedIPsResult, err := inputs.resolveAllowedIP(nil, allowedIPs)

		assert.Nil(t, err)
		assert.Equal(t, []realm.AllowedIP{allowedIPs[0]}, allowedIPsResult)
	})

	for _, tc := range []struct {
		description        string
		selectedAllowedIPs []string
		expectedOutput     []realm.AllowedIP
	}{
		{
			description:        "allow single selection",
			selectedAllowedIPs: []string{"192.1.1.1"},
			expectedOutput:     []realm.AllowedIP{allowedIPs[1]},
		},
		{
			description:        "allow multiple selections",
			selectedAllowedIPs: []string{"0.0.0.0", "192.1.1.1", "192.158.1.38"},
			expectedOutput:     []realm.AllowedIP{allowedIPs[0], allowedIPs[1], allowedIPs[2]},
		},
	} {
		t.Run("should prompt for allowed ips with no input: "+tc.description, func(t *testing.T) {
			inputs := deleteInputs{}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				console.ExpectString("Which IP Address(es) and/or CIDR block(s) would you like to delete?")
				for _, selected := range tc.selectedAllowedIPs {
					console.Send(selected)
					console.Send(" ")
				}
				console.SendLine("")
				console.ExpectEOF()
			}()

			allowedIPsResult, err := inputs.resolveAllowedIP(ui, allowedIPs)

			console.Tty().Close()
			<-doneCh

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, allowedIPsResult)
		})
	}
}
