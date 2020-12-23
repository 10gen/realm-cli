package mock

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewProfile returns a new CLI profile with a random name
func NewProfile(t *testing.T) *profile.Profile {
	t.Helper()
	profile, err := cli.NewProfile(primitive.NewObjectID().Hex())
	assert.Nil(t, err)
	return profile
}
