package accesslist

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
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
		testNewComment string
	}{
		{
			description:    "should return a successful message for a successful update for address and comment",
			testAddress:    "0.0.0.0",
			testNewAddress: "2.2.2.2",
			testNewComment: "new comment",
		},
		{
			description:    "should return a successful message for an update with only an address",
			testAddress:    "0.0.0.0",
			testNewAddress: "68.192.33.2",
		},
		{
			description:    "should return a successful message for an update with only a comment",
			testAddress:    "0.0.0.0",
			testNewComment: "new comment",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			var updateArgs struct {
				groupID, appID, allowedIPID, newAddress, newComment string
			}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{app}, nil
			}
			realmClient.AllowedIPsFn = func(groupID, appID string) ([]realm.AllowedIP, error) {
				return allowedIPs, nil
			}
			realmClient.AllowedIPUpdateFn = func(groupID, appID, allowedIPID, newAddress, newComment string) error {
				updateArgs = struct {
					groupID, appID, allowedIPID, newAddress, newComment string
				}{groupID, appID, allowedIPID, newAddress, newComment}

				return nil
			}

			cmd := &CommandUpdate{updateInputs{
				cli.ProjectInputs{Project: projectID, App: appID},
				tc.testAddress,
				tc.testNewAddress,
				tc.testNewComment,
			}}

			out, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

			assert.Equal(t, "Successfully updated allowed IP\n", out.String())
			assert.Equal(t, "projectID", updateArgs.groupID)
			assert.Equal(t, "appID", updateArgs.appID)
			assert.Equal(t, "address1", updateArgs.allowedIPID)
			assert.Equal(t, tc.testNewAddress, updateArgs.newAddress)
			assert.Equal(t, tc.testNewComment, updateArgs.newComment)
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
				description: "if there is an issue with finding allowed ips for the app",
				clientSetup: func() realm.Client {
					return mock.RealmClient{
						FindAppsFn: func(filter realm.AppFilter) ([]realm.App, error) {
							return []realm.App{app}, nil
						},
						AllowedIPsFn: func(groupID, appID string) ([]realm.AllowedIP, error) {
							return nil, errors.New("something happened with allowed IPs")
						},
					}
				},
				expectedErr: errors.New("something happened with allowed IPs"),
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

	t.Run("resolving the allowed ip", func(t *testing.T) {
		t.Run("should return the allowed ip if searching by address", func(t *testing.T) {
			inputs := updateInputs{
				Address: "0.0.0.0",
			}
			result, err := inputs.resolveAllowedIP(nil, allowedIPs)
			assert.Nil(t, err)
			assert.Equal(t, allowedIPs[0], result)
		})

		t.Run("should return an error if allowed ip is not found", func(t *testing.T) {
			inputs := updateInputs{
				Address: "1.1.1.1",
			}

			_, err := inputs.resolveAllowedIP(nil, allowedIPs)
			assert.Equal(t, errors.New("unable to find allowed IP: 1.1.1.1"), err)
		})

		t.Run("should return an error if neither new ip nor comment is provided", func(t *testing.T) {
			profile := mock.NewProfile(t)
			inputs := updateInputs{
				Address: "1.1.1.1",
			}

			err := inputs.Resolve(profile, nil)
			assert.Equal(t, errors.New("must set either  --new-ip or  --comment when updating an allowed IP address or CIDR block"), err)
		})

		t.Run("should prompt for address when is not provided", func(t *testing.T) {
			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			procedure := func(c *expect.Console) {
				c.ExpectString("Select an IP address or CIDR block to update")
				c.SendLine("address3")
				c.ExpectEOF()
			}

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				procedure(console)
			}()

			inputs := updateInputs{}

			allowedIPsResult, err := inputs.resolveAllowedIP(ui, allowedIPs)

			console.Tty().Close()
			<-doneCh

			assert.Nil(t, err)
			assert.Equal(t, allowedIPs[2], allowedIPsResult)
		})

	})

}
