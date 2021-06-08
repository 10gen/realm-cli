package logs

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestLogTypes(t *testing.T) {
	for _, tc := range []struct {
		logType  string
		logTypes []string
	}{
		{"", nil},
		{logTypeAuth, []string{realm.LogTypeAuth, realm.LogTypeAPIKey}},
		{logTypeFunction, []string{realm.LogTypeFunction}},
		{logTypePush, []string{realm.LogTypePush}},
		{logTypeService, []string{realm.LogTypeServiceFunction, realm.LogTypeWebhook, realm.LogTypeServiceStreamFunction, realm.LogTypeStreamFunction}},
		{logTypeTrigger, []string{realm.LogTypeAuthTrigger, realm.LogTypeDBTrigger, realm.LogTypeScheduledTrigger}},
		{logTypeGraphQL, []string{realm.LogTypeGraphQL}},
		{logTypeSync, []string{realm.LogTypeSyncConnectionStart, realm.LogTypeSyncConnectionEnd, realm.LogTypeSyncSessionStart, realm.LogTypeSyncSessionEnd, realm.LogTypeSyncClientWrite, realm.LogTypeSyncError, realm.LogTypeSyncOther}},
		{logTypeSchema, []string{realm.LogTypeSchemaAdditiveChange, realm.LogTypeSchemaGeneration, realm.LogTypeSchemaValidation}},
	} {
		t.Run("should find log types for type "+tc.logType, func(t *testing.T) {
			i := listInputs{Types: []string{tc.logType}}

			assert.Equal(t, tc.logTypes, i.logTypes())
		})
	}
}
