package run_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/cmds/run"
	"github.com/nokia/ntt/k3"
	"github.com/stretchr/testify/assert"
)

// TestEnqueueJobs verifies that testscase lists are properly generated, if
// user does not specify testscases explicitly.
func TestEnqueueJobs(t *testing.T) {
	tests := []struct {
		files, ids []string
		jobs       []string
		err        error
	}{
		{
			files: []string{"testdata/suite"},
			jobs:  []string{"test.A", "test.B", "test2.C"},
		},
		{
			files: []string{"testdata/suite/"}, ids: []string{"test.A"},
			jobs: []string{"test.A"},
		},
		{
			files: []string{"testdata/suite/test.ttcn3"},
			jobs:  []string{"test.A", "test.B"},
		},
		{
			files: []string{"testdata/lib"},
			jobs:  []string{"math.Test"},
		},
		{
			// Suite does not have a math.Test. Validity of test
			// ids is checked in worker, though.
			files: []string{"testdata/suite"}, ids: []string{"math.Test"},
			jobs: []string{"math.Test"},
		},
	}

	for _, tt := range tests {
		jobs := make(chan run.Job)
		go func() {
			run.EnqueueJobs(tt.files, tt.ids, jobs)
		}()
		var actual []string
		for job := range jobs {
			actual = append(actual, job.Name)
		}
		assert.Equal(t, tt.jobs, actual)
	}
}

func TestRun(t *testing.T) {
	if k3.Runtime() == "" {
		t.Skip("k3 runtime not found")
	}

	tests := []struct {
		files, ids []string
		jobs       []string
		err        error
	}{
		{
			files: []string{"testdata/suite"},
			jobs: []string{
				"test.A(pass)",
				"test.B(pass)",
				"test2.C(pass)"},
		},
		{
			files: []string{"testdata/suite/"}, ids: []string{"test.A"},
			jobs: []string{"test.A(pass)"},
		},
		{
			files: []string{"testdata/suite/test.ttcn3"},
			jobs:  []string{"test.A(pass)", "test.B(pass)"},
		},
		{
			files: []string{"testdata/lib"},
			jobs:  []string{"math.Test(inconc)"},
		},
		{
			files: []string{"testdata/suite"}, ids: []string{"math.Test"},
			jobs: []string{"math.Test()"},
			err:  run.ErrNoSuchTest,
		},
	}

	for _, tt := range tests {
		jobs := make(chan run.Job)
		results := run.Run(context.Background(), jobs)

		go func() {
			run.EnqueueJobs(tt.files, tt.ids, jobs)
		}()

		var (
			actual []string
			err    error
		)
		for res := range results {
			actual = append(actual, fmt.Sprintf("%s(%s)", res.Name, res.Verdict))
			if res.Err != nil {
				err = res.Err
			}
		}
		assert.Equal(t, tt.jobs, actual)
		if !errors.Is(tt.err, err) {
			t.Errorf("expected error %v, got %v", tt.err, err)
		}
	}
}
