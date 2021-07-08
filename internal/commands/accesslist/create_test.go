package accesslist

import (
	"errors"
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
			description: "Should not prompt for inputs when flags provide the data",
			inputs: createInputs{
				Address: "0.0.0.0",
				Comment: "comment",
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Address: "0.0.0.0", Comment: "comment"}, i)
			},
		},
		{
			description: "Should prompt for address when none provided",
			inputs:      createInputs{},
			procedure: func(c *expect.Console) {
				c.ExpectString("IP Address")
				c.SendLine("0.0.0.0")
				c.ExpectEOF()
			},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Address: "0.0.0.0"}, i)
			},
		},
		{
			description: "Should not prompt for address when allow-all flag set",
			inputs: createInputs{
				AllowAll: true,
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{Address: "0.0.0.0", AllowAll: true}, i)
			},
		},
		{
			description: "Should not prompt for address when use-current flag set",
			inputs: createInputs{
				UseCurrent: true,
			},
			procedure: func(c *expect.Console) {},
			test: func(t *testing.T, i createInputs) {
				assert.Equal(t, createInputs{UseCurrent: true}, i)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			profile := mock.NewProfile(t)

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, tc.inputs)
		})
	}

	t.Run("Should error when more than one address given", func(t *testing.T) {
		t.Run("with both address and allow-all", func(t *testing.T) {
			profile := mock.NewProfile(t)

			inputs := createInputs{Address: "0.0.0.0", AllowAll: true}
			err := inputs.Resolve(profile, nil)

			assert.Equal(t, err, errors.New("must only provide one IP address or CIDR block at a time"))
		})
		t.Run("with both address and use-current", func(t *testing.T) {
			profile := mock.NewProfile(t)

			inputs := createInputs{Address: "0.0.0.0", UseCurrent: true}
			err := inputs.Resolve(profile, nil)

			assert.Equal(t, err, errors.New("must only provide one IP address or CIDR block at a time"))
		})
		t.Run("with both allow-all and use-current", func(t *testing.T) {
			profile := mock.NewProfile(t)

			inputs := createInputs{AllowAll: true, UseCurrent: true}
			err := inputs.Resolve(profile, nil)

			assert.Equal(t, err, errors.New("must only provide one IP address or CIDR block at a time"))
		})
		t.Run("with address, allow-all and use-current", func(t *testing.T) {
			profile := mock.NewProfile(t)

			inputs := createInputs{Address: "0.0.0.0", AllowAll: true, UseCurrent: true}
			err := inputs.Resolve(profile, nil)

			assert.Equal(t, err, errors.New("must only provide one IP address or CIDR block at a time"))
		})
	})
}
