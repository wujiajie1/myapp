package bbolt

import (
	"github.com/coreos"
	"syscall"
)

// fdatasync flushes written data to a file descriptor.
func fdatasync(db *coreos.DB) error {
	return syscall.Fdatasync(int(db.file.Fd()))
}
