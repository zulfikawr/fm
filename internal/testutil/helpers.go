package testutil

import (
	"errors"
	"reflect"
	"regexp"
)

// TB is a subset of testing.TB that we use for helpers to be compatible with rapid.T
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
}

// AssertErrorType checks if an error is of a specific type
func AssertErrorType(t TB, err error, target interface{}, msg string) {
	t.Helper()
	if !errors.As(err, target) {
		t.Errorf("%s: expected error type %T, got %T (%v)", msg, target, err, err)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t TB, expected, actual any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %+v, got %+v", msg, expected, actual)
	}
}

// AssertNoError checks if an error is nil
func AssertNoError(t TB, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

var ansiRegex = regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")

// StripANSI removes ANSI escape codes from a string
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
