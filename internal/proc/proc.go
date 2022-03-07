package proc

import (
	"context"
	"os/exec"
	"syscall"
)

// Command returns a exec.Cmd struct to execute the named program with the given arguments.
// The parent death signal is set to SIGKILL for platforms that support it.
func Command(name string, args ...string) *exec.Cmd {
	return CommandContext(context.Background(), name, args...)
}

// CommandContext is like Command but includes a context.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	setPdeathsig(cmd, syscall.SIGKILL)
	return cmd
}

// Exec executes the named program with the given arguments.
// The parent process will be replaced by the new process, for platforms that support it.
func Exec(name string, args ...string) error {
	return execute(name, args...)
}
