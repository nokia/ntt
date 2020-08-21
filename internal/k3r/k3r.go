package k3r

import (
	"io"
	"os"
	"os/exec"

	"github.com/nokia/ntt/internal/ntt"
)

type Instance struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	suite  *ntt.Suite
}

func New(suite *ntt.Suite) (*Instance, error) {

	k3r, err := suite.Getenv("K3R")
	if err != nil {
		return nil, err
	}
	if k3r == "" {
		k3r = "k3r"
	}

	env, err := suite.Environ()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(k3r)
	cmd.Env = append(os.Environ(), env...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &Instance{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (inst *Instance) Run(t string) error {
}
