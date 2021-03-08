package report

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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

	r := Report{
		suite: suite,
		Cores: runtime.NumCPU(),
		Loads: loads(),
	}

	r.Name, _ = suite.Name()

	if db != nil {
		r.db = *db
		r.Collection = *NewCollection(r.Name, db.Runs...)
	}

	return &r, nil
}

func (r *Report) Getenv(s string) string {
	env, _ := r.suite.Getenv(s)
	return env
}

func (r *Report) Environ() []string {
	env := os.Environ()
	env2, _ := r.suite.Environ()
	env = append(env, env2...)
	sort.Strings(env)
	return env
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
			lineSep := []byte{'\n'}
			r, err := os.Open(file)
			if err != nil {
				return err
			}
			for {
				c, err := r.Read(buf)
				atomic.AddInt32(&sum, int32(bytes.Count(buf[:c], lineSep)))

				if err != nil {
					if err == io.EOF {
						err = nil
					}
					return err
				}
			}
		})
	}

	return int(sum), g.Wait()
}

func (r *Report) MaxJobs() int {
	return r.db.MaxJobs
}

func (r *Report) MaxLoad() int {
	return r.db.MaxLoad
}

type Collection struct {
	Name string
	runs RunSlice
}

func NewCollection(name string, runs ...results.Run) *Collection {
	c := &Collection{
		Name: name,
		runs: make(RunSlice, len(runs)),
	}

	for i, r := range runs {
		c.runs[i] = Run{r}
	}
	return c
}

func (c Collection) Runs() RunSlice {
	return c.runs
}

func (c Collection) Tests() RunSlice {
	return NewRunSlice(results.FinalVerdicts(c.runs.asResultsRun()))
}

func (c Collection) Modules() []*Collection {
	suites := make(map[string]*Collection)
	for _, r := range c.runs {
		c, ok := suites[r.Module()]
		if !ok {
			c = NewCollection(r.Module())
		}
		c.runs = append(c.runs, r)
		suites[r.Module()] = c
	}

	ret := make([]*Collection, 0, len(suites))
	for _, v := range suites {
		ret = append(ret, v)
	}
	return ret
}

type RunSlice []Run

func NewRunSlice(runs []results.Run) RunSlice {
	ret := make(RunSlice, len(runs))
	for i, r := range runs {
		ret[i] = Run{r}
	}
	return ret
}

func (rs RunSlice) Total() time.Duration {
	return results.Total(rs.asResultsRun())
}

func (rs RunSlice) Duration() time.Duration {
	return results.Duration(rs.asResultsRun())
}

func (rs RunSlice) First() Run {
	return Run{results.First(rs.asResultsRun())}
}

func (rs RunSlice) Last() Run {
	return Run{results.Last(rs.asResultsRun())}
}

func (rs RunSlice) Shortest() Run {
	return Run{results.Shortest(rs.asResultsRun())}
}

func (rs RunSlice) Longest() Run {
	return Run{results.Longest(rs.asResultsRun())}
}

func (rs RunSlice) Average() time.Duration {
	return results.Average(results.Durations(rs.asResultsRun()))
}

func (rs RunSlice) Deviation() time.Duration {
	return results.Deviation(results.Durations(rs.asResultsRun()))
}

func (rs RunSlice) NotPassed() []Run {
	return rs.filter(func(s string) bool { return s != "pass" })
}

func (rs RunSlice) Failed() []Run {
	return rs.filter(func(s string) bool { return s != "pass" && s != "unstable" })
}

func (rs RunSlice) Unstable() []Run {
	return rs.filter(func(s string) bool { return s == "unstable" })
}

func (rs RunSlice) Result() string {
	switch {
	case len(rs) == 0:
		return "NORUN"

	case len(rs.Failed()) != 0:
		return "FAILED"

	case len(rs.Unstable()) != 0:
		return "UNSTABLE"

	case len(rs.Failed()) == 0 && len(rs.Unstable()) == 0:
		return "PASSED"

	default:
		return "FAILED"
	}
}

func (rs RunSlice) filter(f func(s string) bool) []Run {
	ret := make([]Run, 0, len(rs))
	for _, r := range rs {
		if f(r.Verdict) {
			ret = append(ret, r)
		}
	}
	return ret
}

func (rs RunSlice) asResultsRun() []results.Run {
	ret := make([]results.Run, len(rs))
	for i, r := range rs {
		ret[i] = r.Run
	}
	return ret
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

func (r Run) ReasonFiles() ([]File, error) {
	paths, err := filepath.Glob(filepath.Join(r.WorkingDir, "*.reason"))
	if err != nil {
		return nil, err
	}

	var files []File
	for _, p := range paths {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		files = append(files, File{p, string(b)})
	}
	return files, err
}

type File struct {
	Name    string
	Content string
}
