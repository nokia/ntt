package proc

import (
	"os/exec"
	"syscall"
)

func setPdeathsig(cmd *exec.Cmd, sig syscall.Signal) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: sig,
	}
}
