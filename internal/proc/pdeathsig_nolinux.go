// +build !linux

package proc

import (
	"os/exec"
)

func setPdeathsig(cmd *exec.Cmd) {
}
