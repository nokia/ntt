// +build unix linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func Execute(path string, args ...string) error {
	argv := append([]string{path}, args...)
	return unix.Exec(path, argv, os.Environ())
}
