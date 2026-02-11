package testutil

import (
	"reflect"
	"strings"
	"testing"
)

// AssertEqual verifies that two values are equal.
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %+v (type %T), got %+v (type %T)", msg, expected, expected, actual, actual)
	}
}

// AssertNoError verifies that an error is nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertFatalError verifies that an error is nil, failing fatally if not.
func AssertFatalError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected fatal error: %v", msg, err)
	}
}

// AssertError verifies that an error is NOT nil.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", msg)
	}
}

// AssertErrorContains verifies that an error message contains a specific substring.
func AssertErrorContains(t *testing.T, err error, substr string, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error containing %q, got nil", msg, substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("%s: error %q does not contain %q", msg, err.Error(), substr)
	}
}

// AssertTrue verifies that a condition is true.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true, got false", msg)
	}
}

// AssertFalse verifies that a condition is false.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false, got true", msg)
	}
}

// AssertErrorType verifies that an error is of a specific type.
func AssertErrorType(t *testing.T, err error, target interface{}, msg string) {
	t.Helper()
	if !reflect.TypeOf(err).AssignableTo(reflect.TypeOf(target).Elem()) {
		t.Errorf("%s: expected error type %T, got %T", msg, target, err)
	}
}

// Fail is a helper to fail a test with a formatted message.
func Fail(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Errorf(format, args...)
}
