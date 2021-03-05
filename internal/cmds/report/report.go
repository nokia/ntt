package report

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/results"
	"golang.org/x/sync/errgroup"
)

type Report struct {
	Collection
	suite *ntt.Suite
	db    results.DB
	Cores int
	Loads [3]float64
}

func NewReport(suite *ntt.Suite) (*Report, error) {
	db, err := suite.LatestResults()
	if err != nil {
		return nil, err
	}

	name, _ := suite.Name()

	return &Report{
		Collection: *NewCollection(name, results.FinalVerdicts(db.Runs)...),
		suite:      suite,
		db:         *db,
		Cores:      runtime.NumCPU(),
		Loads:      loads(),
	}, nil
}

func (r *Report) LineCount() (int, error) {
	files, err := r.suite.Files()
	if err != nil {
		return -1, err
	}

	var sum int32
	g := new(errgroup.Group)
	for _, file := range files {
		file := file // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			var err error

			buf := make([]byte, 32*1024)
			count := int32(0)
			lineSep := []byte{'\n'}
			r, err := os.Open(file)
			if err != nil {
				return err
			}
			for {
				c, err := r.Read(buf)
				atomic.AddInt32(&count, int32(bytes.Count(buf[:c], lineSep)))

				if err != nil {
					if err == io.EOF {
						err = nil
					}
					return err
				}
			}

			return nil
		})
	}

	return int(sum), g.Wait()
}

func (r *Report) Runs() *Collection {
	name, _ := r.suite.Name()
	return NewCollection(name, r.db.Runs...)
}

type Collection struct {
	Name  string
	Tests []Run
}

func NewCollection(name string, runs ...results.Run) *Collection {
	c := &Collection{
		Name:  name,
		Tests: make([]Run, len(runs)),
	}

	for i, r := range runs {
		c.Tests[i] = Run{r}
	}
	return c
}

func (c *Collection) runs() []results.Run {
	ret := make([]results.Run, len(c.Tests))
	for i, r := range c.Tests {
		ret[i] = r.Run
	}
	return ret
}

func (c *Collection) Total() time.Duration {
	return results.Total(c.runs())
}

func (c *Collection) Duration() time.Duration {
	return results.Duration(c.runs())
}

func (c *Collection) Shortest() Run {
	return Run{results.Shortest(c.runs())}
}

func (c *Collection) Longest() Run {
	return Run{results.Longest(c.runs())}
}

func (c *Collection) Average() time.Duration {
	return results.Average(results.Durations(c.runs()))
}

func (c *Collection) Deviation() time.Duration {
	return results.Deviation(results.Durations(c.runs()))
}

func (c Collection) Modules() []*Collection {
	suites := make(map[string]*Collection)
	for _, r := range c.Tests {
		c, ok := suites[r.Module()]
		if !ok {
			c = NewCollection(r.Module())
			suites[r.Module()] = c
		}
		c.Tests = append(c.Tests, r)
	}

	ret := make([]*Collection, 0, len(suites))
	for _, v := range suites {
		ret = append(ret, v)
	}
	return ret
}

func (c Collection) NotPassed() []Run {
	return c.filter(func(s string) bool { return s != "pass" })
}

func (c Collection) Failed() []Run {
	return c.filter(func(s string) bool { return s != "pass" && s != "unstable" })
}

func (c Collection) Unstable() []Run {
	return c.filter(func(s string) bool { return s == "unstable" })
}

func (c Collection) filter(f func(s string) bool) []Run {
	ret := make([]Run, 0, len(c.Tests))
	for _, r := range c.Tests {
		if f(r.Verdict) {
			ret = append(ret, r)
		}
	}
	return ret
}

func (c Collection) Result() string {
	switch {
	case len(c.Tests) == 0:
		return "NORUN"

	case len(c.Failed()) != 0:
		return "FAILED"

	case len(c.Unstable()) != 0:
		return "UNSTABLE"

	case len(c.Failed()) == 0 && len(c.Unstable()) == 0:
		return "PASSED"

	default:
		return "FAILED"
	}
}

type Run struct{ results.Run }

func (r Run) Module() string {
	if f := strings.Split(r.Name, "."); len(f) == 2 {
		return f[0]
	}
	return ""
}

func (r Run) Testcase() string {
	f := strings.Split(r.Name, ".")
	return f[len(f)-1]
}
