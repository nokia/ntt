package k3s

import (
	"os/exec"
	"syscall"
)

func setPdeathsig(cmd *exec.Cmd) {
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
}
