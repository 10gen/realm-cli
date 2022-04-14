package realm

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAppResolveProducts(t *testing.T) {
	t.Run("should return the list of default products when none are provided", func(t *testing.T) {
		var emptyArr []string
		for _, tc := range [][]string{
			nil,
			[]string{},
			emptyArr,
		} {
			assert.Equal(t, defaultProducts, resolveProducts(tc))
		}
	})

	t.Run("should return the provided list of products", func(t *testing.T) {
		arr := []string{"one", "two"}
		for _, tc := range [][]string{
			defaultProducts,
			{"standard", "atlas", "charts"},
			arr,
		} {
			assert.Equal(t, tc, resolveProducts(tc))
		}
	})
}

func TestFilterIsEmpty(t *testing.T) {
	for _, tc := range []struct {
		description   string
		filter        AppFilter
		expectedEmpty bool
	}{
		{
			description:   "should return true if struct is empty",
			expectedEmpty: true,
		},
		{
			description:   "should return true if only products has values",
			filter:        AppFilter{Products: []string{"a", "b", "c"}},
			expectedEmpty: true,
		},
		{
			description: "should return false if groupID has value",
			filter:      AppFilter{GroupID: "groupID"},
		},
		{
			description: "should return false if app has value",
			filter:      AppFilter{App: "new-app-abcde"},
		},
		{
			description: "should return false if struct is filled",
			filter:      AppFilter{GroupID: "groupID", App: "new-app-abcde", Products: []string{"a", "b", "c"}},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedEmpty, tc.filter.IsEmpty())
		})
	}
}
