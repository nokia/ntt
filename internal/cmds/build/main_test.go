package build

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/nokia/ntt/internal/log"
)

func init() {
	wrapCmd := func(cmd string) string {
		return fmt.Sprintf("%s -- %s", os.Args[0], cmd)
	}

	os.Setenv("ASN1C", wrapCmd("asn1c"))
	os.Setenv("K3C", wrapCmd("mtc"))
	os.Setenv("CC", wrapCmd("cc"))
	os.Setenv("CXX", wrapCmd("cxx"))
	os.Setenv("ASN2TTCN", wrapCmd("asn1tottcn3"))
	os.Setenv("ASN1CFLAGS", "")
	os.Setenv("K3CFLAGS", "")
	os.Setenv("CFLAGS", "")
	os.Setenv("CXXFLAGS", "")

}

func TestImports(t *testing.T) {
	os.Setenv("NTT_WANT_HELPER_PROCESS", "1")
	defer os.Unsetenv("NTT_WANT_HELPER_PROCESS")

	tests := []struct {
		path   string
		result []string
		err    error
	}{
		{path: "./testdata/invalid/notexist", err: os.ErrNotExist},
		{path: "./testdata/invalid/file.ttcn3", err: syscall.ENOTDIR},
		{path: "./testdata/invalid/dirs", err: ErrNoSources},
		{path: "./testdata/other", err: ErrNoSources},
		{path: "./testdata/🤔", result: []string{"testdata/🤔/a.ttcn3"}},
		{path: "./testdata/lib", result: []string{"testdata/lib/a.ttcn3", "testdata/lib/b.ttcn3", "testdata/lib/🤔.ttcn3"}},
	}

	for _, tt := range tests {
		result, err := buildImport(tt.path)
		if !errors.Is(err, tt.err) {
			t.Errorf("%v: %v, want %v", tt.path, err, tt.err)
		}
		if !equal(result, tt.result) {
			t.Errorf("%v: %v, want %v", tt.path, result, tt.result)
		}
	}

}

// equal returns true if a and b are equal string slices, order is ignored.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for i := range a {
		m[a[i]]++
		m[b[i]]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func TestMain(m *testing.M) {
	if os.Getenv("NTT_WANT_HELPER_PROCESS") != "1" {
		os.Exit(m.Run())
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "TestMain: No command: os.Args=%v\n", os.Args)
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	log.Debugf("TestMain: Command: %v, Args: %v\n", cmd, args)

	switch cmd {
	case "asn1c":
		f := flag.NewFlagSet(cmd, flag.ContinueOnError)
		prefix := f.String("prefix", "", "")
		f.String("output", "", "")
		f.Bool("per", false, "")
		f.Bool("uper", false, "")
		if err := f.Parse(args); err != nil {
			fmt.Fprintf(os.Stderr, "TestMain: Error parsing flags: %v\n", err.Error())
			os.Exit(2)
		}
		log.Debugln("XXXX ASN1C:", f.Args())
		log.Debugln("     prefix:", *prefix)
	default:
	}
}
