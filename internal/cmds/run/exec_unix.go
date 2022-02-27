// +build unix linux

package run

import (
	"os"

	"golang.org/x/sys/unix"
)

func Exec(path string, args ...string) error {
	argv := append([]string{path}, args...)
	return unix.Exec(path, argv, os.Environ())
}
