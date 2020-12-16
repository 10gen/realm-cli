package create

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestUserType(t *testing.T) {
	for _, tc := range []userType{
		// add all user types here
		userTypeAPIKey,
		userTypeEmailPassword,
	} {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidUserType(tc), "must be valid user type")
		})
	}

	t.Run("Should have the correct type representation", func(t *testing.T) {
		assert.Equal(t, "string", userTypeAPIKey.Type())
	})

	t.Run("Should set its value correctly with a valid user type", func(t *testing.T) {
		tc := newUserType()

		assert.Nil(t, tc.ut.Set("email"))
		assert.Equal(t, "email", tc.ut.String())

		assert.Nil(t, tc.ut.Set(""))
		assert.Equal(t, "", tc.ut.String())
	})

	t.Run("Should return an error when setting its value with an invalid user type", func(t *testing.T) {
		tc := newUserType()
		assert.Equal(t, errors.New("unsupported value, use one of [api-key, email] instead"), tc.ut.Set("eggcorn"))
	})
}

type userTypeHolder struct {
	ut *userType
}

func newUserType() userTypeHolder {
	var ut userType
	return userTypeHolder{&ut}
}
