// +build windows

package tail

import (
	"os"
	"vendor"
)

func OpenFile(name string) (file *os.File, err error) {
	return vendor.OpenFile(name, os.O_RDONLY, 0)
}
