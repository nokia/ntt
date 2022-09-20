package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/project"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestJobQueue(t *testing.T) {
	defaultConfig := testConfig(`
module m1 {
    testcase tc1() {}
    testcase tc2() {}
    control {}
}
module m2 {
    testcase tc1() {}
    testcase tc2() {}
    control {}{
}`)

	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		got, err := testJobQueue(t, &project.Config{})
		assert.Nil(t, err)
		assert.Nil(t, got)
	})
	t.Run("default", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig)
		want := []string{
			"m1.control",
			"m2.control",
		}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "all controls are default")
	})
	t.Run("allTests", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "--all-tests")
		want := []string{
			"m1.tc1",
			"m1.tc2",
			"m2.tc1",
			"m2.tc2",
		}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "all tests if requested")
	})
	t.Run("ids", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "m2.tc1", "m2.control")
		want := []string{"m2.tc1", "m2.control"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "only specified tests")
	})
	t.Run("ids", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "--all-tests",
			"m2.tc1",
			"m2.control",
		)
		want := []string{
			"m2.tc1",
			"m2.control",
		}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "only specified tests")
	})
	t.Run("ids", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "", "m2.m2")
		want := []string{"", "m2.m2"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "invalid test ids are bypassed, as they will be checked later by the runner")
	})
	t.Run("tests-file", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-t", "testdata/TestJobQueue.txt")
		want := []string{"m1.tc2"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "only tests from input file are used")
	})
	t.Run("tests-file", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-t", "testdata/TestJobQueue.txt", "m1.tc1")
		want := []string{"m1.tc2", "m1.tc1"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "input test comes before cmd line tests")
	})
	t.Run("tests-file", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-t", "xxx", "m1.tc1")
		assert.True(t, errors.Is(err, os.ErrNotExist))
		assert.Nil(t, got, "no tests at all on error")
	})
	// test basket is only for defaults
	// test basket errors
}

func testConfig(modules ...string) *project.Config {
	conf := &project.Config{}
	for i, m := range modules {
		name := fmt.Sprintf("test://testConfig_%d.ttcn3", i)
		fs.SetContent(name, []byte(m))
		conf.Sources = append(conf.Sources, name)
	}
	return conf
}

func testJobQueue(t *testing.T, conf *project.Config, args ...string) ([]string, error) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.AddFlagSet(RunCommand.Flags())

	// AddFlagSet does not re-initialize the flag values. As a result, flag
	// arguments from earlier tests accumulate and mess up our tests.
	flags.VisitAll(func(f *pflag.Flag) {
		if val, ok := f.Value.(pflag.SliceValue); ok {
			_ = val.Replace(nil)
		}
	})

	if err := flags.Parse(args); err != nil {
		t.Fatal(err)
	}
	allTests, err := flags.GetBool("all-tests")
	if err != nil {
		t.Fatal(err)
	}
	files, err := flags.GetStringSlice("tests-file")
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := JobQueue(context.Background(), flags, conf, files, flags.Args(), allTests)
	var ret []string
	if err == nil {
		for job := range jobs {
			ret = append(ret, job.Name)
		}
	}
	return ret, err
}
