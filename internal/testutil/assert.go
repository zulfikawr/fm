package testutil

import (
	"reflect"
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
