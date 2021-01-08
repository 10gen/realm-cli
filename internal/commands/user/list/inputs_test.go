package list

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestProviderTypeResolution(t *testing.T) {
	for _, tc := range []struct {
		description string
		inputs      inputs
		expectedErr error
	}{
		{
			description: "No provider types should not return an error",
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
					App: "app",
				},
			},
			expectedErr: nil,
		},
		{
			description: "Valid provider types should not return an error",
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
					App: "app",
				},
				ProviderTypes: []string{
					providerTypeLocalUserPass,
					providerTypeAPIKey,
					providerTypeFacebook},
			},
			expectedErr: nil,
		},
		{
			description: "An invalid provider should return an error",
			inputs: inputs{
				ProjectAppInputs: cli.ProjectAppInputs{
					App: "app",
				},
				ProviderTypes: []string{
					providerTypeLocalUserPass,
					"eggcorn",
					providerTypeFacebook},
			},
			expectedErr: errInvalidProviderType,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)
			assert.Equal(t, tc.expectedErr, tc.inputs.Resolve(profile, nil))
		})
	}

}
