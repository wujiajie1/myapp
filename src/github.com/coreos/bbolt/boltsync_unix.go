// +build !windows,!plan9,!linux,!openbsd

package bbolt

import "github.com/coreos"

// fdatasync flushes written data to a file descriptor.
func fdatasync(db *coreos.DB) error {
	return db.file.Sync()
}
