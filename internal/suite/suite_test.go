package suite

// TODO(5nord) implement below tests (shame on me)

// k3 .
// k3 a.ttcn3 b.ttcn
// k3 a.ttcn3 b.ttcn3 c.asn
// k3 . a.ttcn3
// k3 a.ttcn3 .
// k3 xxxx .
// k3 fake.ttcn3 bar.ttcn3
// k3

// source_dir=. k3
// source_dir=. k3 ..
// source_dir=. k3 a.ttcn3

// ttcn3_files=. k3
// ttcn3_files=. k3 ..
// ttcn3_files=. k3 a.ttcn3

// sources=... k3 .
// sources=... k3 a.ttcn3

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuiteFieldValues(t *testing.T) {
	os.Setenv("NTT_TIMEOUT", "")
	s, err := NewFromDirectory("testdata")

	if err != nil {
		t.Fatalf("Error from NewFromdirectory:%s", err.Error())
	}
	assert.Equal(t, 30.4, s.Timeout)
}

func TestSuiteFieldValuesOverwriteFromEnv(t *testing.T) {
	os.Setenv("NTT_TIMEOUT", "12.3")
	s, err := NewFromDirectory("testdata")

	if err != nil {
		t.Fatalf("Error from NewFromdirectory:%s", err.Error())
	}
	assert.Equal(t, 12.3, s.Timeout)
}

func TestSuiteFieldValuesWoYml(t *testing.T) {
	os.Setenv("NTT_TIMEOUT", "")
	// no data will be loaded from package.yml
	var args []string = nil
	s, err := NewFromFiles(args)
	if err != nil {
		t.Fatalf("Error from NewFromdirectory:%s", err.Error())
	}
	assert.True(t, math.IsNaN(s.Timeout))
}

func TestSuiteFieldValuesOverwriteFromEnvWoYml(t *testing.T) {
	os.Setenv("NTT_TIMEOUT", "12.3")
	var args []string = nil
	s, err := NewFromFiles(args)
	if err != nil {
		t.Fatalf("Error from NewFromdirectory:%s", err.Error())
	}
	assert.Equal(t, 12.3, s.Timeout)
}
