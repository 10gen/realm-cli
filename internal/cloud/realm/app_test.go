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
