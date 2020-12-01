// This file is taken from https://github.com/mongodb/mongo-go-driver
// and its `internal/testutil/assert` package.

package assert

import (
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var cmpOpts sync.Map
var errorCompareFn = func(e1, e2 error) bool {
	if e1 == nil || e2 == nil {
		return e1 == nil && e2 == nil
	}

	return e1.Error() == e2.Error()
}
var errorCompareOpts = cmp.Options{cmp.Comparer(errorCompareFn)}

// RegisterOpts registers go-cmp options for a given type
// to be used when calling cmp.Diff or cmp.Equal
func RegisterOpts(t reflect.Type, opts ...cmp.Option) {
	cmpOpts.Store(t, cmp.Options(opts))
}

// Equal compares x and y for equality
// and fails with the test if not
func Equal(t testing.TB, x, y interface{}) {
	t.Helper()
	switch x.(type) {
	case string:
		Equalf(t, x, y, "expected %q to equal %q", y, x)
	default:
		Equalf(t, x, y, "expected %T{%v} to equal %T{%v}", y, y, x, x)
	}
}

// Equalf compares x and y for equality
// and fails with the test with the provided formatted message if not
func Equalf(t testing.TB, x, y interface{}, format string, args ...interface{}) {
	t.Helper()
	if !cmp.Equal(x, y, getCmpOpts(x)...) {
		t.Fatalf(format, args...)
	}
}

// NotEqual compares x and y for inequality
// and fails with the test if not
func NotEqual(t testing.TB, x, y interface{}, format string, args ...interface{}) {
	t.Helper()
	NotEqualf(t, x, y, "expected %T{%v} to not equal %T{%v}", y, y, x, x)
}

// NotEqualf compares x and y for inequality
// and fails with the test with the provided formatted message if not
func NotEqualf(t testing.TB, x, y interface{}, format string, args ...interface{}) {
	t.Helper()
	if cmp.Equal(x, y, getCmpOpts(x)...) {
		t.Fatalf(format, args...)
	}
}

// Match compares x and y and ensures there is no diff
// and fails with the test with the reported differences if not
func Match(t testing.TB, x, y interface{}) {
	t.Helper()
	if diff := cmp.Diff(x, y, getCmpOpts(x)...); diff != "" {
		t.Fatalf(diff)
	}
}

// True asserts that the obj parameter is a boolean with value true
// and fails with the test with the provided formatted message if not
func True(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	b, ok := o.(bool)
	if !ok || !b {
		t.Fatalf(format, args...)
	}
}

// False asserts that the o parameter is a boolean with value false
// and fails with the test with the provided formatted message if not
func False(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	b, ok := o.(bool)
	if !ok || b {
		t.Fatalf(format, args...)
	}
}

// Nil asserts that the o parameter is nil
// and fails with the test if not
func Nil(t testing.TB, o interface{}) {
	t.Helper()
	Nilf(t, o, "expected %T to be <nil>, but it was not: %v", o, o)
}

// Nilf asserts that the o parameter is nil
// and fails with the test with the provided formatted message if not
func Nilf(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	if !isNil(o) {
		t.Fatalf(format, args...)
	}
}

// NotNil asserts that the o parameter is not nil
// and fails with the test if not
func NotNil(t testing.TB, o interface{}) {
	t.Helper()
	NotNilf(t, o, "expected %T to be not <nil>, but it was", o)
}

// NotNilf asserts that the o parameter is not nil
// and fails with the test with the provided formatted message if not
func NotNilf(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	if isNil(o) {
		t.Fatalf(format, args...)
	}
}

func getCmpOpts(o interface{}) cmp.Options {
	if opts, ok := cmpOpts.Load(reflect.TypeOf(o)); ok {
		return opts.(cmp.Options)
	}
	if _, ok := o.(error); ok {
		return errorCompareOpts
	}
	return nil
}

func isNil(o interface{}) bool {
	if o == nil {
		return true
	}

	val := reflect.ValueOf(o)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
