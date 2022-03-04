package run_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/run"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	t3xf := testBuild(t)

	tests := []struct {
		input   string
		timeout time.Duration
		events  []string
	}{
		{
			input: "test.A",
			events: []string{
				`tciTestCaseStarted test.A`,
				`tciTestCaseTerminated test.A pass`,
			}},
		{
			input: "test.control",
			events: []string{
				`tciControlStarted test.control`,
				`tciTestCaseStarted test.B`,
				`tciTestCaseTerminated test.B fail`,
				`tciTestCaseStarted test.A`,
				`tciTestCaseTerminated test.A pass`,
				`tciControlTerminated test.control pass`,
			}},
		{
			input: "test2.control",
			events: []string{
				`tciControlStarted test2.control`,
				`tciTestCaseStarted test2.A`,
				`tciTestCaseTerminated test2.A pass`,
				`tciControlTerminated test2.control pass`,
			}},
		{
			input: "test3.control",
			events: []string{
				`tciControlStarted test3.control`,
				`tciControlTerminated test3.control pass`, // no error message when control does not exist
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
				"tciError error (id not fully qualified)",
			}},
		{
			input: "test3",
			events: []string{
				"tciError error (id not fully qualified)",
			}},
		{
			input: "control",
			events: []string{
				"tciError error (id not fully qualified)",
			}},
		{
			input:   "math.Test",
			timeout: 1 * time.Second,
			events: []string{
				`tciTestCaseStarted math.Test`,
				`tciError error (timeout)`,
			}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, time.Duration(tt.timeout))
			}

			var actual []string
			for e := range run.NewTest(t3xf, tt.input).RunWithContext(ctx) {
				actual = append(actual, e.String())
			}
			assert.Equal(t, tt.events, actual)
			if cancel != nil {
				cancel()
			}
		})
	}

}

func testBuild(t *testing.T) string {
	t.Helper()
	if k3r := k3.Runtime(); k3r == "k3r" {
		t.Skip("no k3 runtime found")
	}

	suite, err := ntt.NewFromArgs("testdata/suite")
	if err != nil {
		t.Fatal(err)
	}
	name, err := suite.Name()
	if err != nil {
		t.Fatal(err)
	}

	if err := ntt2.BuildProject(name, suite); err != nil {
		t.Fatal(err)
	}
	return cache.Lookup(fmt.Sprintf("%s.t3xf", name))
}
