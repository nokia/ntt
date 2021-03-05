package results

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// A DB struct contains results of last test executable execution
type DB struct {
	Version int
	MaxJobs int
	MaxLoad int
	Runs    []Run
}

// A Run describes the execution of a single test case.
type Run struct {
	Name     string // Full qualified test name
	Instance int    // Test instance
	Verdict  string // the test verdict (pass, fail, none, ...)
	Reason   string // Optional reason for verdicts

	Begin Timestamp // When the test was started
	End   Timestamp // When the test ended

	WorkingDir string  // Working Directory of the test
	Load       float64 // the system load when the test was started
	MaxMem     int     // the maximum memory used when the test ended

	RunnerID string `json:"runnerid"`
}

// A unique identifier of the run. Usually something like "testname-2"
func (r Run) ID() string {
	return r.Name + "-" + fmt.Sprintf("%d\n", r.Instance)
}

// Duration of an individual run
func (r Run) Duration() time.Duration {
	return r.End.Sub(r.Begin.Time)
}

// String returns a printable and simplified representation of Run
func (r Run) String() string {
	return fmt.Sprintf("%s	%s	%s", r.Verdict, r.ID(), r.Duration())
}

// Timestamp is a Unix timestamp in milliseconds.
type Timestamp struct {
	time.Time
}

// MarshalJSON is used to convert the timestamp to JSON
func (t Timestamp) MarshalJSON() ([]byte, error) {
	ms := t.UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(ms, 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *Timestamp) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	secs := q / 1000
	ms := q % 1000
	t.Time = time.Unix(secs, ms*int64(time.Millisecond))
	return nil
}

// Total returns total time spent if all runs would have been executed sequentially.
func Total(runs []Run) time.Duration {
	var sum time.Duration
	for _, r := range runs {
		sum += r.Duration()
	}
	return sum
}

// Duration returns the time spent between the first and the last test run.
func Duration(runs []Run) time.Duration {
	if len(runs) == 0 {
		return 0
	}
	return Last(runs).End.Sub(First(runs).Begin.Time)
}

// First returns the first test run
func First(runs []Run) Run {
	if len(runs) == 0 {
		return Run{}
	}

	first := runs[0]
	for _, r := range runs {
		if r.Begin.Before(first.Begin.Time) {
			first = r
		}
	}
	return first
}

// Last returns the last test run
func Last(runs []Run) Run {
	if len(runs) == 0 {
		return Run{}
	}

	last := runs[0]
	for _, r := range runs {
		if r.End.After(last.End.Time) {
			last = r
		}
	}
	return last
}

// Shortest returns the shortest test run
func Shortest(runs []Run) Run {
	if len(runs) == 0 {
		return Run{}
	}

	shortest := runs[0]
	for _, r := range runs {
		if r.Duration() < shortest.Duration() {
			shortest = r
		}
	}
	return shortest
}

// Longest returns the longest test run
func Longest(runs []Run) Run {
	if len(runs) == 0 {
		return Run{}
	}

	longest := runs[0]
	for _, r := range runs {
		if r.Duration() > longest.Duration() {
			longest = r
		}
	}
	return longest
}

// Duration returns a slice of test run durations
func Durations(runs []Run) []time.Duration {
	ret := make([]time.Duration, len(runs))
	for i := range runs {
		ret[i] = runs[i].Duration()
	}
	return ret
}

// FinalVerdicts folds multiple runs instances into one with a final verdict.
//
// * A test is considered "pass", if all runs had verdict "pass".
// * A test is considered "unstable", if only some runs had verdict "pass"
// * Runs with worse or equal verdict will overwrite previous runs.
func FinalVerdicts(runs []Run) []Run {

	severity := func(verdict string) int {
		switch strings.TrimSpace(strings.ToLower(verdict)) {
		case "pass":
			return 0
		case "none", "":
			return 1
		case "inconc":
			return 2
		case "fail":
			return 3
		case "error":
			return 4
		default:
			return 5
		}
	}

	var (
		tests    = make(map[string]Run)
		names    = make([]string, 0, len(runs))
		passed   = make(map[string]bool)
		unstable = make(map[string]bool)
	)

	for _, run := range runs {
		last, ok := tests[run.Name]
		if !ok {
			names = append(names, run.Name)
			last = run
		}

		// A test is a success if it passes at least once.
		if run.Verdict == "pass" {
			passed[run.Name] = true
		}

		// A test is unstable if it has its verdicts change
		if run.Verdict != last.Verdict {
			unstable[run.Name] = true
		}

		// Remember worst and latest test run for post analysis
		if severity(run.Verdict) >= severity(last.Verdict) {
			tests[run.Name] = run
		}
	}

	ret := make([]Run, len(names))
	for i, name := range names {
		t := tests[name]
		if passed[name] {
			t.Verdict = "pass"
			if unstable[name] {
				t.Verdict = "unstable"
			}
		}

		ret[i] = t
	}
	return ret
}

// Average returns the average test duration (median)
func Average(slice []time.Duration) time.Duration {
	if len(slice) == 0 {
		return 0
	}

	sort.Sort(durationSlice(slice))

	n := len(slice) / 2

	// odd slice length
	if len(slice)&1 == 1 {
		return slice[n]
	}

	// even slice length
	return (slice[n-1] + slice[n]) / 2

}

// Mean returns arithmetic mean
func Mean(slice []time.Duration) time.Duration {
	var sum int
	for _, d := range slice {
		sum += int(d)
	}

	return time.Duration(sum / len(slice))
}

// Deviation returns the standard deviation of duration
func Deviation(slice []time.Duration) time.Duration {
	if len(slice) == 0 {
		return 0
	}

	mean := Mean(slice)

	v := 0.0
	for _, d := range slice {
		v += math.Pow(float64(d), 2) / float64(mean)
	}

	return time.Duration(math.Sqrt(v))
}

type durationSlice []time.Duration

func (a durationSlice) Len() int           { return len(a) }
func (a durationSlice) Less(i, j int) bool { return a[i] < a[j] }
func (a durationSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
