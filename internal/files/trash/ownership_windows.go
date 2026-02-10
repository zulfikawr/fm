//go:build windows

package trash

import (
	"os"
)

// getOwnership returns 0, 0 on Windows as UID/GID are not supported in this context.
func getOwnership(_ os.FileInfo) (uid, gid int) {
	return 0, 0
}
