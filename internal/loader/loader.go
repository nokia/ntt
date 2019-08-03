package loader

import (
	"sort"

	"github.com/nokia/ntt/internal/loc"
	st "github.com/nokia/ntt/internal/suite"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type Suite struct {
	*st.Suite
	ParseOnly bool
	Fset      *loc.FileSet
	Modules   []*ast.Module
}

func NewFromArgs(args []string) (*Suite, error) {
	var err error
	suite := &Suite{}
	suite.Suite, err = st.NewFromArgs(args)
	if err != nil {
		return nil, err
	}
	suite.SetEnv()
	sort.Strings(suite.Sources)
	return suite, nil
}

func (suite *Suite) Load() error {
	suite.Fset = loc.NewFileSet()
	m, err := parseFiles(suite.Fset, suite.Sources)
	if err != nil {
		return err
	}

	suite.Modules = m
	return nil
}
