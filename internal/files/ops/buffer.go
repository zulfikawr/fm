package ops

import (
	"sync"

	"github.com/zulfikawr/fm/internal/constants"
)

// bufferPool reuses memory buffers for I/O operations to reduce GC pressure
var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, constants.CopyBufferSize)
		return &b
	},
}

// GetBuffer acquires a buffer from the pool
func GetBuffer() []byte {
	return *bufferPool.Get().(*[]byte)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf []byte) {
	// Zero out the buffer if needed or just return it
	// For file copying, zeroing is usually not necessary as it will be overwritten
	bufferPool.Put(&buf)
}
