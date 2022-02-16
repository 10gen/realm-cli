package feedback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestErrNew(t *testing.T) {
	t.Run("should create an error with minimal details", func(t *testing.T) {
		err := NewErr(errors.New("something bad happened"))

		assert.Equal(t, cliErr{cause: errors.New("something bad happened")}, err)

		t.Run("and be recognized as a usage hider", func(t *testing.T) {
			var usageHider ErrUsageHider
			assert.True(t, errors.As(err, &usageHider), fmt.Sprintf("expected error to be ErrUsageHider, but got %T instead", err))
			assert.False(t, usageHider.HideUsage(), "should not hide usage")
		})

		t.Run("and be recognized as a suggester", func(t *testing.T) {
			var suggester ErrSuggester
			assert.True(t, errors.As(err, &suggester), fmt.Sprintf("expected error to be ErrSuggester, but got %T instead", err))
			assert.True(t, len(suggester.Suggestions()) == 0, "should have no suggestions")
		})

		t.Run("and be recognized as a link referrer", func(t *testing.T) {
			var linkReferrer ErrLinkReferrer
			assert.True(t, errors.As(err, &linkReferrer), fmt.Sprintf("expected error to be ErrLinkReferrer, but got %T instead", err))
			assert.True(t, len(linkReferrer.ReferenceLinks()) == 0, "should have no links")
		})
	})

	t.Run("should create an error with complete details", func(t *testing.T) {
		err := NewErr(
			errors.New("something bad happened"),
			ErrNoUsage{}, ErrSuggestion{"suggestion"}, ErrReferenceLink{"ref.link"},
		)

		assert.Equal(t,
			cliErr{
				cause:          errors.New("something bad happened"),
				hideUsage:      true,
				suggestions:    []interface{}{"suggestion"},
				referenceLinks: []interface{}{"ref.link"},
			},
			err,
		)

		t.Run("and be recognized as a usage hider", func(t *testing.T) {
			var usageHider ErrUsageHider
			assert.True(t, errors.As(err, &usageHider), fmt.Sprintf("expected error to be ErrUsageHider, but got %T instead", err))
			assert.True(t, usageHider.HideUsage(), "should hide usage")
		})

		t.Run("and be recognized as a suggester", func(t *testing.T) {
			var suggester ErrSuggester
			assert.True(t, errors.As(err, &suggester), fmt.Sprintf("expected error to be ErrSuggester, but got %T instead", err))
			assert.Equal(t, suggester.Suggestions(), []interface{}{"suggestion"})
		})

		t.Run("and be recognized as a link referrer", func(t *testing.T) {
			var linkReferrer ErrLinkReferrer
			assert.True(t, errors.As(err, &linkReferrer), fmt.Sprintf("expected error to be ErrLinkReferrer, but got %T instead", err))
			assert.Equal(t, linkReferrer.ReferenceLinks(), []interface{}{"ref.link"})
		})
	})
}

func TestErrWrap(t *testing.T) {
	t.Run("should inherit all of the wrapped error details", func(t *testing.T) {
		assert.Equal(t,
			cliErr{
				cause:          errors.New("failed to process: something bad happened"),
				hideUsage:      true,
				suggestions:    []interface{}{"suggestion"},
				referenceLinks: []interface{}{"ref.link"},
			},
			WrapErr("failed to process: %w", NewErr(
				errors.New("something bad happened"),
				ErrNoUsage{}, ErrSuggestion{"suggestion"}, ErrReferenceLink{"ref.link"},
			)),
		)
	})

	t.Run("should be able to layer additional error details on top of the wrapped details", func(t *testing.T) {
		assert.Equal(t,
			cliErr{
				cause:          errors.New("failed to process: something bad happened"),
				hideUsage:      false,
				suggestions:    []interface{}{"a better suggestion", "suggestion"},
				referenceLinks: []interface{}{"a better ref.link", "ref.link"},
			},
			WrapErr("failed to process: %w",
				NewErr(
					errors.New("something bad happened"),
					ErrNoUsage{}, ErrSuggestion{"suggestion"}, ErrReferenceLink{"ref.link"},
				),
				errWithUsage{}, ErrSuggestion{"a better suggestion"}, ErrReferenceLink{"a better ref.link"},
			),
		)
	})

	t.Run("should work without the wrapped fmt verb", func(t *testing.T) {
		assert.Equal(t,
			cliErr{cause: errors.New("failed to process: something bad happened")},
			WrapErr("failed to process: %s", NewErr(errors.New("something bad happened"))),
		)
	})
}

// needed for testing purposes only
type errWithUsage struct{}

func (err errWithUsage) ApplyTo(details *ErrDetails) {
	var b bool
	details.HideUsage = &b
}
