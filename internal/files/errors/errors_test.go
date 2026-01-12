package errors

import (
	"errors"
	"fm/internal/testutil"
	"fmt"
	"pgregory.net/rapid"
	"testing"
)

func TestErrorWrapping_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 5: Error Wrapping Consistency
	rapid.Check(t, func(t *rapid.T) {
		op := rapid.String().Draw(t, "op")
		path := rapid.String().Draw(t, "path")
		errMsg := rapid.String().Draw(t, "errMsg")
		err := errors.New(errMsg)

		wrapped := WrapErrorWithPath(err, op, path)

		var fe *FileError
		testutil.AssertErrorType(t, wrapped, &fe, "Wrapped error")

		if fe.Op != op {
			t.Errorf("Expected op %s, got %s", op, fe.Op)
		}

		if fe.Path != path {
			t.Errorf("Expected path %s, got %s", path, fe.Path)
		}

		if !errors.Is(wrapped, err) {
			t.Errorf("Wrapped error should wrap the original error")
		}
	})
}

func TestUnsupportedOperationError_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 4: Unsupported Operation Error Type
	rapid.Check(t, func(t *rapid.T) {
		op := rapid.String().Draw(t, "op")
		fs := rapid.String().Draw(t, "fs")

		err := &UnsupportedOperationError{Op: op, Filesystem: fs}
		wrapped := WrapErrorWithPath(err, "GeneralOp", "some/path")

		var fe *FileError
		testutil.AssertErrorType(t, wrapped, &fe, "Wrapped error")

		var ue *UnsupportedOperationError
		testutil.AssertErrorType(t, fe.Err, &ue, "Underlying error")

		expectedMsg := fmt.Sprintf("operation %s is not supported by %s filesystem", op, fs)
		if fe.Msg != expectedMsg {
			t.Errorf("Expected message %s, got %s", expectedMsg, fe.Msg)
		}
	})
}
