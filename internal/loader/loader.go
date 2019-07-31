package loader

import (
	st "github.com/nokia/ntt/internal/suite"
	"github.com/nokia/ntt/internal/ttcn3/syntax"
)

type Suite struct {
	*st.Suite
	ParseOnly bool
	Fset      *syntax.FileSet
	Modules   []*syntax.Module
}

func NewFromArgs(args []string) (*Suite, error) {
	var err error
	suite := &Suite{}
	suite.Suite, err = st.NewFromArgs(args)
	if err != nil {
		return nil, err
	}
	suite.SetEnv()
	return suite, nil
}

func (suite *Suite) Load() error {
	suite.Fset = syntax.NewFileSet()
	m, err := parseFiles(suite.Fset, suite.Sources)
	if err != nil {
		return err
	}

	suite.Modules = m
	return nil
}
