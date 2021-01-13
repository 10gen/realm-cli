package shared

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestProviderType(t *testing.T) {
	for _, tc := range ValidProviderTypes {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidProviderType(tc), "must be valid provider type")
		})
	}
}

func TestStatusType(t *testing.T) {
	for _, tc := range ValidStatusTypes {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidStatusType(tc), "must be valid status type")
		})
	}
}

func TestUserStateType(t *testing.T) {
	for _, tc := range ValidUserStateTypes {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidUserStateType(tc), "must be valid user state type")
		})
	}
}
