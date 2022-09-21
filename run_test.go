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

    // @wip
    control {}
}
module m2 {
    testcase tc1() {}
    testcase tc2() {}
    control {}
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
	t.Run("tests-file", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-t", "testdata/TestJobQueueEmpty.txt")
		assert.Nil(t, err)
		assert.Nil(t, got, "an empty file means no tests will be run")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-a", "-r", "tc1")
		want := []string{"m1.tc1", "m2.tc1"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "default tests are filtered")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-a", "-r", "xxx")
		assert.Nil(t, err)
		assert.Nil(t, got, "it's not an error if no tests match")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-r", "tc1", "tc1", "tc2")
		want := []string{"tc1"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "CLI tests are filtered")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-a", "-x", "tc1")
		want := []string{"m1.tc2", "m2.tc2"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "default tests are filtered")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-x", "tc1", "tc1")
		assert.Nil(t, err)
		assert.Nil(t, got, "it's not an error all tests match")
	})
	t.Run("filters", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-x", "xxx", "tc1")
		want := []string{"tc1"}
		assert.Nil(t, err)
		assert.Equal(t, want, got, "it's not an error if no tests match")
	})
	t.Run("tags", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-R", "@wip")
		want := []string{"m1.control"}
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("tags", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-X", "@wip")
		want := []string{"m2.control"}
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("tags", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-R", "@wip", "m1.control", "m2.control", "foo")
		want := []string{"m1.control"}
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("tags", func(t *testing.T) {
		got, err := testJobQueue(t, defaultConfig, "-X", "@wip", "m2.control", "foo")
		want := []string{"m2.control", "foo"}
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
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

	var (
		allTests bool
		files    []string
	)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.AddFlagSet(BasketFlags())
	flags.BoolVarP(&allTests, "all-tests", "a", false, "run all tests instead of control parts")
	flags.StringSliceVarP(&files, "tests-file", "t", nil, "read tests from FILE. If this option is used multiple times all contained tests will be executed in that order. When FILE is '-', read standard input")
	if err := flags.Parse(args); err != nil {
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
