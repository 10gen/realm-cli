package user

import (
	"testing"

	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProfileResolveFlags(t *testing.T) {
	t.Run("should provide defaults if flags are empty and set them in the profile", func(t *testing.T) {
		profile, err := NewProfile(primitive.NewObjectID().Hex())
		assert.Nil(t, err)

		assert.Equal(t, telemetry.ModeEmpty, profile.Flags.TelemetryMode)
		assert.Equal(t, "", profile.Flags.RealmBaseURL)
		assert.Equal(t, "", profile.Flags.AtlasBaseURL)

		assert.Nil(t, profile.ResolveFlags())

		assert.Equal(t, telemetry.ModeEmpty, profile.Flags.TelemetryMode)
		assert.Equal(t, telemetry.ModeEmpty, profile.TelemetryMode())

		assert.Equal(t, defaultRealmBaseURL, profile.Flags.RealmBaseURL)
		assert.Equal(t, defaultRealmBaseURL, profile.RealmBaseURL())

		assert.Equal(t, defaultAtlasBaseURL, profile.Flags.AtlasBaseURL)
		assert.Equal(t, defaultAtlasBaseURL, profile.AtlasBaseURL())
	})

	t.Run("should use flags to set them in the profile", func(t *testing.T) {
		profile, err := NewProfile(primitive.NewObjectID().Hex())
		assert.Nil(t, err)

		profile.Flags = Flags{
			TelemetryMode: telemetry.ModeStdout,
			RealmBaseURL:  "https://realm-dev.mongodb.com",
			AtlasBaseURL:  "https://cloud-dev.mongodb.com",
		}

		assert.Equal(t, telemetry.ModeStdout, profile.Flags.TelemetryMode)
		assert.Equal(t, "https://realm-dev.mongodb.com", profile.Flags.RealmBaseURL)
		assert.Equal(t, "https://cloud-dev.mongodb.com", profile.Flags.AtlasBaseURL)

		assert.Nil(t, profile.ResolveFlags())

		assert.Equal(t, telemetry.ModeStdout, profile.Flags.TelemetryMode)
		assert.Equal(t, telemetry.ModeStdout, profile.TelemetryMode())

		assert.Equal(t, "https://realm-dev.mongodb.com", profile.Flags.RealmBaseURL)
		assert.Equal(t, "https://realm-dev.mongodb.com", profile.RealmBaseURL())

		assert.Equal(t, "https://cloud-dev.mongodb.com", profile.Flags.AtlasBaseURL)
		assert.Equal(t, "https://cloud-dev.mongodb.com", profile.AtlasBaseURL())
	})
}
