package run_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/cmds/build"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/run"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	t3xf := testBuild(t)

	tests := []struct {
		input   string
		timeout float64
		events  []string
	}{
		{
			input: "test.A",
			events: []string{
				`tciTestCaseStarted "test.A"`,
				`tciTestCaseTerminated pass`,
			}},
		{
			input: "test.control",
			events: []string{
				`tciTestCaseStarted "test.B"`,
				`tciTestCaseTerminated fail`,
				`tciTestCaseStarted "test.A"`,
				`tciTestCaseTerminated pass`,
				`tciControlTerminated`,
			}},
		{
			input: "test2.control",
			events: []string{
				`tciTestCaseStarted "test2.A"`,
				`tciTestCaseTerminated pass`,
				`tciControlTerminated`,
			}},
		{
			input: "test3.control",
			events: []string{
				`tciControlTerminated`, // no error message when control does not exist
			}},
		{
			input: "test3.X",
			events: []string{
				"tciError error (no such test case)",
			}},
		{
			input: "X.X",
			events: []string{
				"tciError error (no such module)",
			}},
		{
			input: "test3.test3",
			events: []string{
				"tciError error (exit status 2)", // Exit 2, due to exception.
			}},
		{
			input: "asd",
			events: []string{
				"tciError error (module not ready)",
			}},
		{
			input: "test3",
			events: []string{
				"tciError error (module not ready)",
			}},
		{
			input: "control",
			events: []string{
				"tciError error (module not ready)",
			}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			var actual []string
			for e := range run.NewTest(t3xf, tt.input).Run() {
				actual = append(actual, e.String())
			}
			assert.Equal(t, tt.events, actual)
		})
	}

}

func testBuild(t *testing.T) string {
	t.Helper()
	if k3r := k3.Runtime(); k3r == "k3r" {
		t.Skip("no k3 runtime found")
	}

	args := []string{"testdata/suite"}
	build.Command.SetArgs(args)
	if err := build.Command.Execute(); err != nil {
		t.Fatal(err)
	}

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		t.Fatal(err)
	}
	name, err := suite.Name()
	if err != nil {
		t.Fatal(err)
	}
	return cache.Lookup(fmt.Sprintf("%s.t3xf", name))
}
