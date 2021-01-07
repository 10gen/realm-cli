package list

import (
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProviderTypesValue(t *testing.T) {
	for _, tc := range []string{
		providerTypeLocalUserPass,
		providerTypeAPIKey,
		providerTypeFacebook,
		providerTypeGoogle,
		providerTypeAnonymous,
		providerTypeCustom,
	} {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidProviderType(tc), "must be valid provider type")
		})
	}

	t.Run("Should have the correct type representation", func(t *testing.T) {
		pt := []string{}
		assert.Equal(t, "stringSlice", newProviderTypesValue(&pt).Type())
	})

	t.Run("Should set its value correctly with a valid provider type", func(t *testing.T) {
		pt := []string{}
		ptv := newProviderTypesValue(&pt)

		assert.Nil(t, ptv.Set(""))
		assert.Equal(t, "[]", ptv.String())

		assert.Nil(t, ptv.Set(providerTypeAPIKey))
		assert.Equal(t, "[api-key]", ptv.String())

	})

	t.Run("Should return an error when setting its value with an invalid provider type", func(t *testing.T) {
		pt := []string{}
		ptv := newProviderTypesValue(&pt)
		allProviderTypes := []string{
			providerTypeLocalUserPass,
			providerTypeAPIKey,
			providerTypeFacebook,
			providerTypeGoogle,
			providerTypeAnonymous,
			providerTypeCustom,
		}
		assert.Equal(t, fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allProviderTypes, ", ")), ptv.Set("eggcorn"))
	})
}
