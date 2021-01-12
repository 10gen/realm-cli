package shared

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProviderType(t *testing.T) {
	for _, tc := range ValidProviderTypes {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, IsValidProviderType(tc), "must be valid provider type")
		})
	}
}
