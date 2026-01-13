package ops

import (
	"context"
	"io"
)

type cancellableReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *cancellableReader) Read(p []byte) (n int, err error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(p)
	}
}

type cancellableWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (w *cancellableWriter) Write(p []byte) (n int, err error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
		return w.writer.Write(p)
	}
}

// NewCancellableReader wraps a reader with context cancellation support
func NewCancellableReader(ctx context.Context, r io.Reader) io.Reader {
	return &cancellableReader{ctx: ctx, reader: r}
}

// NewCancellableWriter wraps a writer with context cancellation support
func NewCancellableWriter(ctx context.Context, w io.Writer) io.Writer {
	return &cancellableWriter{ctx: ctx, writer: w}
}
