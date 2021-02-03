package user

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestCreateInputs(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      createInputs
		procedure   func(c *expect.Console)
		test        func(t *testing.T, i createInputs)
	}{
		{
			description: "with user type set to email and email and password flags not set",
			inputs:      createInputs{UserType: userTypeEmailPassword},
			procedure: func(c *expect.Console) {
				c.ExpectString("Email")
				c.SendLine("user@domain.com")
				c.ExpectString("Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "user@domain.com", i.Email)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "with user type set to email and only password flag not set",
			inputs:      createInputs{UserType: userTypeEmailPassword, Email: "user@domain.com"},
			procedure: func(c *expect.Console) {
				c.ExpectString("Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "user@domain.com", i.Email)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "with user type set to email and only email flag not set",
			inputs:      createInputs{UserType: userTypeEmailPassword, Password: "password"},
			procedure: func(c *expect.Console) {
				c.ExpectString("Email")
				c.SendLine("user@domain.com")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "user@domain.com", i.Email)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "with user type set to apiKey and api key name flag not set",
			inputs:      createInputs{UserType: userTypeAPIKey},
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key Name")
				c.SendLine("publickey")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "publickey", i.APIKeyName)
			},
		},
		{
			description: "with user type set to apiKey and api key name flag set",
			inputs:      createInputs{UserType: userTypeAPIKey, APIKeyName: "publickey"},
			procedure: func(c *expect.Console) {
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "publickey", i.APIKeyName)
			},
		},
		{
			description: "with no user type set but email flag is set",
			inputs:      createInputs{Email: "user@domain.com"},
			procedure: func(c *expect.Console) {
				c.ExpectString("Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "user@domain.com", i.Email)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "with no user type set but api key flag is set",
			inputs:      createInputs{APIKeyName: "publickey"},
			procedure: func(c *expect.Console) {
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "publickey", i.APIKeyName)
			},
		},
		{
			description: "with no flags set should be able to set details for a new email password user",
			procedure: func(c *expect.Console) {
				c.ExpectString("Which auth provider type are you creating a user for?")
				c.SendLine("em") // type enough to filter, then hit enter
				c.ExpectString("New Email")
				c.SendLine("user@domain.com")
				c.ExpectString("New Password")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "user@domain.com", i.Email)
				assert.Equal(t, "password", i.Password)
			},
		},
		{
			description: "with no flags set should be able to set details for a new api key user",
			procedure: func(c *expect.Console) {
				c.ExpectString("Which auth provider type are you creating a user for?")
				c.SendLine("ap") // type enough to filter, then hit enter
				c.ExpectString("API Key Name")
				c.SendLine("publickey")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, "publickey", i.APIKeyName)
			},
		},
	} {
		t.Run(fmt.Sprintf("%s setup should prompt for the missing inputs", tc.description), func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			tc.inputs.App = "some-app" // avoid app resolution
			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}
}
