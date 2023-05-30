package control

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
)

var ErrNoSuch = errors.New("no such")

// NewTestPlan parses the TTCN-3 sources provided in the given project
// configuration and crestes an empty test plan. Syntax errors in the TTCN-3
// source files are ignored. The given configuration must not be nil.
func NewTestPlan(conf *project.Config) (*TestPlan, error) {
	if conf == nil {
		panic("NewTestPlan: project.Config must not be nil")
	}

	srcs, err := fs.TTCN3Files(conf.Sources...)
	if err != nil {
		return nil, err
	}

	tp := &TestPlan{
		m:    sync.Map{},
		conf: conf,
	}

	tests := make([][]string, len(srcs))
	controls := make([][]string, len(srcs))
	wg := sync.WaitGroup{}
	wg.Add(len(srcs))
	start := time.Now()
	for i, src := range srcs {
		go func(src string, i int) {
			defer wg.Done()
			var (
				mod         string
				modLvl, lvl int
			)
			root := ttcn3.ParseFile(src)
			root.Inspect(func(n syntax.Node) bool {

				// We need to keep track of the current module
				// to be able to generate fully qualified
				// testcase names.
				if n == nil {
					if lvl == modLvl {
						mod = ""
						modLvl = 0
					}
					lvl--
				} else {
					lvl++
				}

				switch n := n.(type) {
				case *syntax.Module:
					mod = n.Name.String()
					modLvl = lvl
					return true
				case *syntax.FuncDecl:
					switch {
					case n.IsTest():
						name := ttcn3.JoinNames(mod, n.Name.String())
						tp.m.Store(name, n)
						tests[i] = append(tests[i], name)
					case n.IsControl():
						name := ttcn3.JoinNames(mod, n.Name.String())
						tp.m.Store(name, n)
						controls[i] = append(controls[i], name)

					}
					return false
				case *syntax.ControlPart:
					name := ttcn3.JoinNames(mod, n.Name.String())
					tp.m.Store(name, n)
					controls[i] = append(controls[i], name)
					return false

				default:
					return true
				}
			})
		}(src, i)
	}
	wg.Wait()
	log.Debugf("Scanned all tests in %s.\n", time.Since(start))

	for _, t := range tests {
		tp.Tests = append(tp.Tests, t...)
	}
	for _, c := range controls {
		tp.Controls = append(tp.Controls, c...)
	}
	return tp, nil
}

// A TestPlan is a ordered collection of test cases and runtime parameters.
type TestPlan struct {
	conf *project.Config
	m    sync.Map

	// Controls is a ordered list of fully qualified control functions.
	Controls []string

	// Tests is a ordered list of fully qualified test case names.
	Tests []string
}

// Next returns the next Job to be executed. If there are no more jobs, nil is
// returned.
func (tp *TestPlan) Next() *Job {
	return nil
}

// Add adds the given test case to the test plan.
func (tp *TestPlan) Add(name string) error {
	return fmt.Errorf("%w: %s", ErrNoSuch, name)
}
