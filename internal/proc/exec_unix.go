// +build unix linux

package proc

import (
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func execute(name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return err
	}
	argv := append([]string{name}, args...)
	return unix.Exec(path, argv, os.Environ())
}
