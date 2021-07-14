package accesslist

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestAllowedIPUpdateHandler(t *testing.T) {
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

	for _, tc := range []struct {
		description    string
		testAddress    string
		testNewAddress string
		testComment    string
	}{
		{
			description:    "should return a successful message for a successful update for address and comment",
			testAddress:    "0.0.0.0",
			testNewAddress: "2.2.2.2",
			testComment:    "new comment",
		},
		{
			description:    "should return a successful message for an update with only an address",
			testAddress:    "192.1.1.1",
			testNewAddress: "68.192.33.2",
		},
		{
			description: "should return a successful message for an empty update",
			testAddress: "192.158.1.38",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}
			realmClient.AllowedIPsFn = func(groupID, appID string) ([]realm.AllowedIP, error) {
				return allowedIPs, nil
			}
			realmClient.AllowedIPUpdateFn = func(groupID, appID, allowedIPID, newAddress, comment string) error {
				return nil
			}

			cmd := &CommandUpdate{updateInputs{
				cli.ProjectInputs{projectID, appID, nil},
				tc.testAddress,
				tc.testNewAddress,
				tc.testComment,
			}}

			out, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, "Successfully updated allowed IP\n", out.String())
		})
	}

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      updateInputs
			clientSetup func() realm.Client
			expectedErr error
		}{
			{
				description: "if there is no app",
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return nil, errors.New("Something went wrong with the app")
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return allowedIPs, nil
						},
					}
				},
				expectedErr: errors.New("Something went wrong with the app"),
			},
			{
				description: "if there is an issue with finding allowed ips for the app",
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return nil, errors.New("Something happened with allowed IPs")
						},
					}
				},
				expectedErr: errors.New("Something happened with allowed IPs"),
			},
			{
				description: "if there is an issue with finding the allowed ip specified in the access list",
				inputs:      updateInputs{Address: "1.1.1.1"},
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return allowedIPs, nil
						},
					}
				},
				expectedErr: errors.New("unable to find allowed IP: 1.1.1.1"),
			},
			{
				description: "if there is an issue with updating the allowed ip",
				inputs:      updateInputs{Address: allowedIPs[0].Address},
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return allowedIPs, nil
						},
						AllowedIPUpdateFn: func(groupID, appID, allowedIPID, newAddress, comment string) error {
							return errors.New("something bad happened")
						},
					}
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				_, ui := mock.NewUI()

				realmClient := tc.clientSetup()
				cmd := &CommandUpdate{tc.inputs}
				assert.Equal(t, tc.expectedErr, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			})
		}
	})
}

func TestAllowedIPUpdateInputs(t *testing.T) {
	allowedIPs := []realm.AllowedIP{
		{ID: "address1", Address: "0.0.0.0", Comment: "comment"},
		{ID: "address2", Address: "192.1.1.1"},
		{ID: "address3", Address: "192.158.1.38", Comment: "cool comment"},
	}

	t.Run("should return the allowed ip if searching by address", func(t *testing.T) {
		inputs := updateInputs{
			Address: "0.0.0.0",
		}
		result, err := inputs.resolveAllowedIP(nil, allowedIPs)
		assert.Nil(t, err)
		assert.Equal(t, allowedIPs[0], result)
	})

	t.Run("should return an error if we cannot find the allowed ip to update", func(t *testing.T) {
		inputs := updateInputs{
			Address: "1.1.1.1",
		}

		_, err := inputs.resolveAllowedIP(nil, allowedIPs)
		assert.Equal(t, errors.New("unable to find allowed IP: 1.1.1.1"), err)
	})

	t.Run("should show the prompt for a user if the input is empty", func(t *testing.T) {
		inputs := updateInputs{}

		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		doneCh := make(chan struct{})
		go func() {
			defer close(doneCh)
			console.ExpectString("Which IP address or CIDR block would you like to update?")
			console.SendLine("address3")
			console.ExpectEOF()
		}()

		allowedIPsResult, err := inputs.resolveAllowedIP(ui, allowedIPs)

		console.Tty().Close()
		<-doneCh

		assert.Nil(t, err)
		assert.Equal(t, allowedIPs[2], allowedIPsResult)
	})
}
