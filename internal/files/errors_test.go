package files

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		op   string
		want string
	}{
		{
			"nil error",
			nil,
			"Op",
			"",
		},
		{
			"not exist",
			os.ErrNotExist,
			"Stat",
			"Stat failed: file or directory does not exist",
		},
		{
			"permission denied",
			os.ErrPermission,
			"Open",
			"Open failed: permission denied",
		},
		{
			"already exists",
			os.ErrExist,
			"Create",
			"Create failed: destination already exists",
		},
		{
			"disk full",
			errors.New("no space left on device"),
			"Write",
			"Write failed: disk is full",
		},
		{
			"EOF",
			io.EOF,
			"Read",
			"Read failed: unexpected end of file",
		},
		{
			"generic error",
			errors.New("something went wrong"),
			"Op",
			"Op failed: something went wrong",
		},
		{
			"double wrap",
			WrapError(os.ErrNotExist, "Stat"),
			"Op",
			"Stat failed: file or directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapError(tt.err, tt.op)
			if tt.err == nil {
				if got != nil {
					t.Errorf("WrapError() = %v, want nil", got)
				}
				return
			}
			if got.Error() != tt.want {
				t.Errorf("WrapError() = %q, want %q", got.Error(), tt.want)
			}
		})
	}
}

func TestFileErrorUnwrap(t *testing.T) {
	orig := os.ErrNotExist
	wrapped := WrapError(orig, "Stat")

	if !errors.Is(wrapped, orig) {
		t.Errorf("Expected wrapped error to be orig error")
	}

	var fe *FileError
	if !errors.As(wrapped, &fe) {
		t.Errorf("Expected wrapped error to be *FileError")
	} else if fe.Op != "Stat" {
		t.Errorf("Expected Op to be Stat, got %s", fe.Op)
	}
}
