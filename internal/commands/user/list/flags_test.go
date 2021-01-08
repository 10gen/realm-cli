package list

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestIsEachProviderTypeValid(t *testing.T) {
	for _, tc := range []struct {
		description     string
		providerTypes   []string
		expectedIsValid bool
	}{
		{
			description:     "No provider types should be valid",
			providerTypes:   []string{},
			expectedIsValid: true,
		},
		{
			description: "Provider types with an invalid type should not be valid",
			providerTypes: []string{
				providerTypeLocalUserPass,
				providerTypeAPIKey,
				"eggcorn",
				providerTypeFacebook,
			},
			expectedIsValid: false,
		},
		{
			description: "All valid provider types should be valid",
			providerTypes: []string{
				providerTypeLocalUserPass,
				providerTypeAPIKey,
				providerTypeFacebook,
				providerTypeGoogle,
				providerTypeAnonymous,
				providerTypeCustom,
			},
			expectedIsValid: true,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsValid, isEachProviderTypeValid(tc.providerTypes))
		})
	}

}
