package accesslist

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestAllowedIPCreateHandler(t *testing.T) {
	projectID := "projectID"
	appID := "appID"
	allowedIPID := "allowedIPID"
	allowedIPAddress := "allowedIPAddress"
	allowedIPComment := "allowedIPComment"
	allowedIPUseCurrent := false
	allowedIPAllowAll := false
	app := realm.App{
		ID:          appID,
		GroupID:     projectID,
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	t.Run("should create an allowed ip", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		var capturedFilter realm.AppFilter
		var capturedGroupID, capturedAppID, capturedIPAddress, capturedComment string
		var capturedUseCurrent bool
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			capturedFilter = filter
			return []realm.App{app}, nil
		}

		realmClient.AllowedIPCreateFn = func(groupID, appID, ipAddress, comment string, useCurrent bool) (realm.AllowedIP, error) {
			capturedGroupID = groupID
			capturedAppID = appID
			capturedIPAddress = ipAddress
			capturedComment = comment
			capturedUseCurrent = useCurrent
			return realm.AllowedIP{allowedIPID, allowedIPAddress, allowedIPComment, allowedIPUseCurrent}, nil
		}

		cmd := &CommandCreate{createInputs{
			ProjectInputs: cli.ProjectInputs{
				Project: projectID,
				App:     appID,
			},
			Address:    allowedIPAddress,
			Comment:    allowedIPComment,
			UseCurrent: allowedIPUseCurrent,
			AllowAll:   allowedIPAllowAll,
		}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, fmt.Sprintf("Successfully created allowed IP, id: %s\n", "allowedIPID"), out.String())

		t.Log("and should properly pass through the expected inputs")
		assert.Equal(t, realm.AppFilter{projectID, appID, nil}, capturedFilter)
		assert.Equal(t, projectID, capturedGroupID)
		assert.Equal(t, appID, capturedAppID)
		assert.Equal(t, allowedIPAddress, capturedIPAddress)
		assert.Equal(t, allowedIPComment, capturedComment)
		assert.Equal(t, allowedIPUseCurrent, capturedUseCurrent)
	})

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
				description: "when creating an allowed IP fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.AllowedIPCreateFn = func(groupID, appID, ipAddress, comment string, useCurrent bool) (realm.AllowedIP, error) {
						return realm.AllowedIP{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandCreate{}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
func TestAllowedIPCreateInputs(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      createInputs
		test        func(t *testing.T, i createInputs)
	}{
		{
			description: "should not prompt for inputs when flags provide the data",
			inputs: createInputs{
				Address: "0.0.0.0",
				Comment: "comment",
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Address: "0.0.0.0", Comment: "comment"}, i)
			},
		},
		{
			description: "should not prompt for address when allow-all flag set",
			inputs: createInputs{
				AllowAll: true,
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Address: "0.0.0.0", AllowAll: true}, i)
			},
		},
		{
			description: "should not prompt for address when use-current flag set",
			inputs: createInputs{
				UseCurrent: true,
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{UseCurrent: true}, i)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)
			assert.Nil(t, tc.inputs.Resolve(profile, nil))

			tc.test(t, tc.inputs)
		})
	}

	t.Run("should prompt for address when none provided", func(t *testing.T) {
		_, console, _, ui, consoleErr := mock.NewVT10XConsole()
		assert.Nil(t, consoleErr)
		defer console.Close()

		profile := mock.NewProfile(t)
		procedure := func(c *expect.Console) {
			c.ExpectString("IP Address")
			c.SendLine("0.0.0.0")
			c.ExpectEOF()
		}

		doneCh := make(chan (struct{}))
		go func() {
			defer close(doneCh)
			procedure(console)
		}()

		inputs := createInputs{}

		assert.Nil(t, inputs.Resolve(profile, ui))

		console.Tty().Close() // flush the writers
		<-doneCh              // wait for procedure to complete

		assert.Equal(t, createInputs{Address: "0.0.0.0"}, inputs)
	})

	t.Run("should error when more than one address given", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			inputs      createInputs
		}{
			{
				description: "with allow all specified and an address provided",
				inputs:      createInputs{Address: "0.0.0.0", AllowAll: true},
			},
			{
				description: "with use current specified and an address provided",
				inputs:      createInputs{Address: "0.0.0.0", UseCurrent: true},
			},
			{
				description: "with both allow all and use current specified",
				inputs:      createInputs{AllowAll: true, UseCurrent: true},
			},
			{
				description: "with both allow all and use current specified and an address provided",
				inputs:      createInputs{Address: "0.0.0.0", AllowAll: true, UseCurrent: true},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile := mock.NewProfile(t)
				err := tc.inputs.Resolve(profile, nil)

				assert.Equal(t, errTooManyAddressess, err)
			})
		}
	})
}
