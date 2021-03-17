package api

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestContentTypeByExtension(t *testing.T) {
	t.Run("should return the matched mime type", func(t *testing.T) {
		for _, tc := range []struct {
			ext      string
			mimeType string
		}{
			{"mp4", "video/mp4"},
			{"mpe", "video/mpeg"},
			{"qtif", "image/x-quicktime"},
			{"mesh", "model/mesh"},
			{"so", "application/octet-stream"},
			{"lzh", "application/octet-stream"},
		} {
			t.Run(fmt.Sprintf("with %s extension", tc.ext), func(t *testing.T) {
				mimeType, ok := ContentTypeByExtension(tc.ext)
				assert.True(t, ok, "should be ok")
				assert.Equal(t, tc.mimeType, mimeType)
			})
		}
	})

	t.Run("should not return a matched mime type", func(t *testing.T) {
		for _, tc := range []string{
			"nat",
			"CPT",
		} {
			t.Run(fmt.Sprintf("with %s extension", tc), func(t *testing.T) {
				_, ok := ContentTypeByExtension(tc)
				assert.False(t, ok, "should not be ok")
			})
		}
	})

}
