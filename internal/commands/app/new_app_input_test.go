package app

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/atlas"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
	"github.com/Netflix/go-expect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppNewAppInputsResolveFrom(t *testing.T) {
	t.Run("should do nothing if from is not set", func(t *testing.T) {
		var i newAppInputs
		f, err := i.resolveFrom(nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, from{}, f)
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
		procedure      func(c *expect.Console)
		findAppErr     error
		groupsErr      error
		expectedFrom   from
		expectedFilter realm.AppFilter
		expectedErr    error
	}{
		{
			description:    "should return the app id and group id of specified app when project and from are set",
			inputs:         newAppInputs{Project: testApp.GroupID, From: testApp.ID},
			procedure:      func(c *expect.Console) {},
			expectedFrom:   from{GroupID: testApp.GroupID, AppID: testApp.ID},
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
		},
		{
			description: "should return the app id and group id of specified app when from is set and project is selected",
			inputs:      newAppInputs{From: testApp.ID},
			procedure: func(c *expect.Console) {
				c.ExpectString("Atlas Project")
				c.Send(testApp.GroupID)
				c.SendLine(" ")
				c.ExpectEOF()
			},
			expectedFrom:   from{GroupID: testApp.GroupID, AppID: testApp.ID},
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
		},
		{
			description:    "should error when finding group",
			inputs:         newAppInputs{From: testApp.ID},
			procedure:      func(c *expect.Console) {},
			groupsErr:      errors.New("atlas client error"),
			expectedFilter: realm.AppFilter{},
			expectedErr:    errors.New("atlas client error"),
		},
		{
			description:    "should error when finding app",
			inputs:         newAppInputs{Project: testApp.GroupID, From: testApp.ID},
			procedure:      func(c *expect.Console) {},
			findAppErr:     errors.New("realm client error"),
			expectedFilter: realm.AppFilter{GroupID: testApp.GroupID, App: testApp.ID},
			expectedErr:    errors.New("realm client error"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var appFilter realm.AppFilter
			rc := mock.RealmClient{}
			rc.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				appFilter = filter
				return []realm.App{testApp}, tc.findAppErr
			}

			ac := mock.AtlasClient{}
			ac.GroupsFn = func() ([]atlas.Group, error) {
				return []atlas.Group{{ID: testApp.GroupID}}, tc.groupsErr
			}

			_, console, _, ui, consoleErr := mock.NewVT10XConsole()
			assert.Nil(t, consoleErr)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			f, err := tc.inputs.resolveFrom(ui, rc, ac)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedFrom, f)
			assert.Equal(t, tc.expectedFilter, appFilter)
		})
	}
}
