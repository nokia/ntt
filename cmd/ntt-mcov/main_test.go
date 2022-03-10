package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/run"
	"github.com/stretchr/testify/assert"
)

func TestMessageCoverage(t *testing.T) {
	expected := []string{
		"test.A.fa1	1",
		"test.A.fa2.fb1	1",
		"test.A.fa2.fb2	0",
		"test.B.fb1	0",
		"test.B.fb2	1",
		"test.MessageA.a	2",
		"test.MessageA.b	0",
		"test.MessageA.c	1",
		"test.MessageB	0",
		"test.MessageC.a	0",
		"test.MessageC.f	1",
	}

	w := bytes.Buffer{}
	mcov(runK3(t), &w)
	assert.Equal(t, expected, strings.Split(strings.TrimSpace(w.String()), "\n"))
}

func runK3(t *testing.T) io.Reader {
	t.Helper()
	old, cancel := initStage(t)
	defer cancel()
	os.Setenv("K3RFLAGS", "--record-fmt=alist")
	t3xf := testBuild(t, filepath.Join(old, "testdata/test.ttcn3"))
	test := run.NewTest(t3xf, "test.control")
	for range test.Run() {
	}
	f, err := os.Open(test.LogFile)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func initStage(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	old, err = filepath.Rel(dir, old)
	if err != nil {
		t.Fatal(err)
	}
	return old, func() {
		os.Chdir(old)
		os.RemoveAll(dir)
	}
}

func testBuild(t *testing.T, args ...string) string {
	t.Helper()
	if k3r := k3.Runtime(); k3r == "k3r" {
		t.Skip("no k3 runtime found")
	}

	suite, err := ntt.NewFromArgs(args...)
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
