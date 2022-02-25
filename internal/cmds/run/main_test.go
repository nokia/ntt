package run_test

import (
	"testing"

	"github.com/nokia/ntt/internal/cmds/run"
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
