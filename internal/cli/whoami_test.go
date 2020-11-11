package cli

import (
	"bytes"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWhoamiHandler(t *testing.T) {
	t.Run("Handler should run as a noop", func(t *testing.T) {
		cmd := &whoamiCommand{}

		err := cmd.Handler(nil, nil, nil)
		u.MustMatch(t, cmp.Diff(nil, err))
	})
}

func TestWhoamiFeedback(t *testing.T) {
	t.Run("Feedback should print the auth details", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			setup       func(t *testing.T, profile *Profile)
			test        func(t *testing.T, output string)
		}{
			{
				description: "with no user logged in",
				test: func(t *testing.T, output string) {
					u.MustMatch(t, cmp.Diff("No user is currently logged in.\n", output))
				},
			},
			{
				description: "with a user fully logged in",
				setup: func(t *testing.T, profile *Profile) {
					profile.SetUser("username", "password")
					profile.SetSession("accessToken", "refreshToken")
				},
				test: func(t *testing.T, output string) {
					// TODO(REALMC-7339): once the table printer is implemented, add tests here asserting as much
				},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
				u.MustMatch(t, cmp.Diff(nil, profileErr))

				if tc.setup != nil {
					tc.setup(t, profile)
				}

				buf := new(bytes.Buffer)
				ui := mock.NewUI(mock.UIOptions{}, buf)

				cmd := &whoamiCommand{}
				err := cmd.Feedback(profile, ui)
				u.MustMatch(t, cmp.Diff(nil, err))

				tc.test(t, buf.String())
			})
		}
	})
}
