package delete

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestStatusType(t *testing.T) {
	for _, tc := range []statusType{
		// add all status types here
		statusTypeConfirmed,
		statusTypePending,
	} {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidStatusType(tc), "must be valid status type")
		})
	}

	t.Run("Should have the correct type representation", func(t *testing.T) {
		assert.Equal(t, "string", statusTypeConfirmed.Type())
	})

	t.Run("Should set its value correctly with a valid status type", func(t *testing.T) {
		tc := newStatusType()

		assert.Nil(t, tc.pt.Set("confirmed"))
		assert.Equal(t, "confirmed", tc.pt.String())

		assert.Nil(t, tc.pt.Set(""))
		assert.Equal(t, "", tc.pt.String())
	})

	t.Run("Should return an error when setting its value with an invalid status type", func(t *testing.T) {
		tc := newStatusType()
		assert.Equal(t, errors.New("unsupported value, use one of [confirmed, pending] instead"), tc.pt.Set("eggcorn"))
	})
}

type statusTypeHolder struct {
	pt *statusType
}

func newStatusType() statusTypeHolder {
	var pt statusType
	return statusTypeHolder{&pt}
}
