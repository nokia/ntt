package lsp

import "github.com/nokia/ntt/project"

type TestController struct {
}

func (ctrl *TestController) IsRunning(p project.Interface, name string) bool {
	return false
}
