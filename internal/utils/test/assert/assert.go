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
func Equal(t testing.TB, expected, actual interface{}) {
	t.Helper()
	switch expected.(type) {
	case string:
		Equalf(t, expected, actual, "failed to assert equals ( actual, expected )\n\t%q\n\t%q", actual, expected)
	default:
		Equalf(t, expected, actual, "failed to assert equals ( actual, expected )\n\t%T{%+v}\n\t%T{%+v}", actual, actual, expected, expected)
	}
}

// Equalf compares x and y for equality
// and fails with the test with the provided formatted message if not
func Equalf(t testing.TB, expected, actual interface{}, format string, args ...interface{}) {
	t.Helper()
	if !cmp.Equal(expected, actual, getCmpOpts(expected)...) {
		t.Fatalf("\n"+format, args...)
	}
}

// NotEqual compares x and y for inequality
// and fails with the test if not
func NotEqual(t testing.TB, expected, actual interface{}, format string, args ...interface{}) {
	t.Helper()
	switch expected.(type) {
	case string:
		NotEqualf(t, expected, actual, "failed to assert not equals ( actual, expected )\n\t%q\n\t%q", actual, expected)
	default:
		NotEqualf(t, expected, actual, "failed to assert not equals ( actual, expected )\n\t%T{%+v}\n\t%T{%+v}", actual, actual, expected, expected)
	}
}

// NotEqualf compares x and y for inequality
// and fails with the test with the provided formatted message if not
func NotEqualf(t testing.TB, expected, actual interface{}, format string, args ...interface{}) {
	t.Helper()
	if cmp.Equal(expected, actual, getCmpOpts(expected)...) {
		t.Fatalf("\n"+format, args...)
	}
}

// Match compares x and y and ensures there is no diff
// and fails with the test with the reported differences if not
func Match(t testing.TB, expected, actual interface{}) {
	t.Helper()
	if diff := cmp.Diff(expected, actual, getCmpOpts(expected)...); diff != "" {
		t.Fatalf("\nfailed to assert no diff:\n" + diff)
	}
}

// True asserts that the obj parameter is a boolean with value true
// and fails with the test with the provided formatted message if not
func True(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	b, ok := o.(bool)
	if !ok || !b {
		t.Fatalf("\n"+format, args...)
	}
}

// False asserts that the o parameter is a boolean with value false
// and fails with the test with the provided formatted message if not
func False(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	b, ok := o.(bool)
	if !ok || b {
		t.Fatalf("\n"+format, args...)
	}
}

// Nil asserts that the o parameter is nil
// and fails with the test if not
func Nil(t testing.TB, o interface{}) {
	t.Helper()
	Nilf(t, o, "failed to assert nil: %T{%+v}", o, o)
}

// Nilf asserts that the o parameter is nil
// and fails with the test with the provided formatted message if not
func Nilf(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	if !isNil(o) {
		t.Fatalf("\n"+format, args...)
	}
}

// NotNil asserts that the o parameter is not nil
// and fails with the test if not
func NotNil(t testing.TB, o interface{}) {
	t.Helper()
	NotNilf(t, o, "failed to assert not nil: %T", o)
}

// NotNilf asserts that the o parameter is not nil
// and fails with the test with the provided formatted message if not
func NotNilf(t testing.TB, o interface{}, format string, args ...interface{}) {
	t.Helper()
	if isNil(o) {
		t.Fatalf("\n"+format, args...)
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
