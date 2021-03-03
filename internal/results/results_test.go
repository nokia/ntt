package results

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func run(verdict, id string) Run {
	fields := strings.Split(id, "-")
	if len(fields) != 2 {
		panic("id must be of format $name-$id")
	}
	name := fields[0]
	inst, err := strconv.Atoi(fields[1])
	if err != nil {
		panic(err)
	}
	return Run{
		Verdict:  verdict,
		Name:     name,
		Instance: inst,
	}
}

func stringSlice(runs []Run) []string {
	ret := make([]string, len(runs))
	for i := range runs {
		ret[i] = runs[i].String()
	}
	return ret
}

func TestFinalVerdict(t *testing.T) {
	actual := stringSlice(FinalVerdicts([]Run{
		run("fail", "Test.A-0"),
		run("pass", "Test.A-1"),
		run("pass", "Test.B-1"),
		run("fail", "Test.B-0"),
		run("pass", "Test.C-0"),
		run("fail", "Test.C-1"),
		run("fail", "Test.D-1"),
		run("pass", "Test.D-0"),
		run("pass", "Test.E-0"),
		run("pass", "Test.E-1"),
		run("pass", "Test.F-1"),
		run("pass", "Test.F-0"),
		run("fail", "Test.G-0"),
		run("fail", "Test.G-1"),
		run("fail", "Test.H-1"),
		run("fail", "Test.H-0"),
		run("none", "Test.I-0"),
		run("inconc", "Test.I-1"),
		run("inconc", "Test.J-1"),
		run("none", "Test.J-0"),
		run("inconc", "Test.K-0"),
		run("none", "Test.K-1"),
		run("none", "Test.L-1"),
		run("inconc", "Test.L-0"),
	}))

	expected := stringSlice([]Run{
		run("unstable", "Test.A-0"),
		run("unstable", "Test.B-0"),
		run("unstable", "Test.C-1"),
		run("unstable", "Test.D-1"),
		run("pass", "Test.E-1"),
		run("pass", "Test.F-0"),
		run("fail", "Test.G-1"),
		run("fail", "Test.H-0"),
		run("inconc", "Test.I-1"),
		run("inconc", "Test.J-1"),
		run("inconc", "Test.K-0"),
		run("inconc", "Test.L-0"),
	})

	assert.Equal(t, expected, actual)
}
