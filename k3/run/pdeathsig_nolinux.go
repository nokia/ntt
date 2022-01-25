// +build !linux

package run

import (
	"os/exec"
)

func setPdeathsig(cmd *exec.Cmd) {
}
