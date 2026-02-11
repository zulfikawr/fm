package errors

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFileError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *FileError
		expected string
	}{
		{
			name: "With path and msg",
			err: &FileError{
				Op:   "Open",
				Path: "/test/path",
				Msg:  "custom message",
			},
			expected: "Open failed at /test/path: custom message",
		},
		{
			name: "With path, no msg, with err",
			err: &FileError{
				Op:   "Open",
				Path: "/test/path",
				Err:  errors.New("original error"),
			},
			expected: "Open failed at /test/path: original error",
		},
		{
			name: "No path, with msg",
			err: &FileError{
				Op:  "Read",
				Msg: "no path here",
			},
			expected: "Read failed: no path here",
		},
		{
			name: "No path, no msg, with err",
			err: &FileError{
				Op:  "Read",
				Err: errors.New("something went wrong"),
			},
			expected: "Read failed: something went wrong",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("FileError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileError_Unwrap(t *testing.T) {
	original := errors.New("original")
	fe := &FileError{Err: original}
	if !errors.Is(fe, original) {
		t.Errorf("Expected fe to wrap original error")
	}
}

func TestOtherErrors(t *testing.T) {
	t.Run("UnsupportedOperationError", func(t *testing.T) {
		err := &UnsupportedOperationError{Op: "Zip", Filesystem: "SFTP"}
		expected := "operation Zip is not supported by SFTP filesystem"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{Field: "Name", Value: 123, Message: "must be string"}
		expected := "validation failed for Name (value: 123): must be string"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("PermissionError", func(t *testing.T) {
		err := &PermissionError{Path: "/root", Operation: "Write"}
		expected := "permission denied for Write on path /root"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})
}

func TestWrapError(t *testing.T) {
	t.Run("Nil error", func(t *testing.T) {
		if got := WrapError(nil, "Op"); got != nil {
			t.Errorf("WrapError(nil) = %v, want nil", got)
		}
	})

	t.Run("WrapErrorWithPath double wrap", func(t *testing.T) {
		inner := &FileError{Op: "Inner", Msg: "msg"}
		outer := WrapErrorWithPath(inner, "Outer", "/some/path")
		fe := outer.(*FileError)
		if fe.Path != "/some/path" {
			t.Errorf("Expected path /some/path, got %s", fe.Path)
		}
	})

	t.Run("Specialized errors", func(t *testing.T) {
		ue := &UnsupportedOperationError{Op: "Zip", Filesystem: "SFTP"}
		err := WrapError(ue, "Op")
		if !strings.Contains(err.Error(), ue.Error()) {
			t.Errorf("Expected wrapped error to contain %q", ue.Error())
		}

		ve := &ValidationError{Field: "F", Value: "V", Message: "M"}
		err = WrapError(ve, "Op")
		if !strings.Contains(err.Error(), ve.Error()) {
			t.Errorf("Expected wrapped error to contain %q", ve.Error())
		}

		pe := &PermissionError{Path: "/P", Operation: "O"}
		err = WrapError(pe, "Op")
		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected wrapped error to contain 'permission denied'")
		}
	})

	t.Run("Common system errors", func(t *testing.T) {
		tests := []struct {
			err      error
			expected string
		}{
			{os.ErrNotExist, "file or directory does not exist or command not found"},
			{os.ErrExist, "destination already exists"},
			{os.ErrPermission, "permission denied"},
			{io.EOF, "unexpected end of file"},
			{fmt.Errorf("context deadline exceeded"), "timed out"},
			{fmt.Errorf("no space left on device"), "disk is full"},
			{fmt.Errorf("text file busy"), "file is currently in use by another process"},
			{fmt.Errorf("not a directory"), "not a directory"},
			{fmt.Errorf("is a directory"), "cannot perform this operation on a directory"},
		}

		for i := range tests {
			tt := tests[i]
			err := WrapError(tt.err, "Op")
			fe := err.(*FileError)
			if fe.Msg != tt.expected {
				t.Errorf("For error %v, expected msg %q, got %q", tt.err, tt.expected, fe.Msg)
			}
		}
	})

	t.Run("Unknown error", func(t *testing.T) {
		unknown := errors.New("unknown error")
		err := WrapError(unknown, "Op")
		fe := err.(*FileError)
		if fe.Err != unknown {
			t.Errorf("Expected wrapped error to be %v, got %v", unknown, fe.Err)
		}
	})
}
