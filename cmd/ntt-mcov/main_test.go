package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/k3r"
	"github.com/nokia/ntt/project"
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
	old := initStage(t)
	os.Setenv("K3RFLAGS", "--record-fmt=alist")
	t3xf := testBuild(t, filepath.Join(old, "testdata/test.ttcn3"))
	test := k3r.NewTest(t3xf, "test.control")
	for range test.Run() {
	}
	f, err := os.Open(test.LogFile)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func initStage(t *testing.T) string {
	dir := t.TempDir()
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

	t.Cleanup(func() {
		os.Chdir(old)
	})

	return old
}

func testBuild(t *testing.T, args ...string) string {
	t.Helper()
	if k3r := k3.Runtime(); k3r == "k3r" {
		t.Skip("no k3 runtime found")
	}

	p, err := project.Open(args...)
	if err != nil {
		t.Fatal(err)
	}
	if err := project.Build(p); err != nil {
		t.Fatal(err)
	}
	return p.K3.T3XF
}
