// +build !linux

package k3s

import (
	"os/exec"
)

func setPdeathsig(cmd *exec.Cmd) {
}
