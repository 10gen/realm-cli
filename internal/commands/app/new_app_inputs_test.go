package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppNewAppInputsResolveFrom(t *testing.T) {
	t.Run("should do nothing if from is not set", func(t *testing.T) {
		var i newAppInputs
		r, err := i.resolveRemoteApp(nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, realm.App{}, r)
	})

	testApp := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "test-app-abcde",
		Name:        "test-app",
	}

	for _, tc := range []struct {
		description    string
		inputs         newAppInputs
		expectedRemote realm.App
		expectedFilter realm.AppFilter
		expectedErr    error
	}{
		{
			description:    "should return the app id and group id of specified app when from is set",
			inputs:         newAppInputs{RemoteApp: testApp.ID},
			expectedRemote: testApp,
			expectedFilter: realm.AppFilter{App: testApp.ID},
		},
		{
			description:    "should error when finding app",
			inputs:         newAppInputs{RemoteApp: testApp.ID},
			expectedFilter: realm.AppFilter{App: testApp.ID},
			expectedErr:    errors.New("realm client error"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter
			rc := mock.RealmClient{}
			rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return []realm.App{testApp}, tc.expectedErr
			}

			r, err := tc.inputs.resolveRemoteApp(nil, rc)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedRemote, r)
			assert.Equal(t, tc.expectedFilter, appFilter)
		})
	}
}
