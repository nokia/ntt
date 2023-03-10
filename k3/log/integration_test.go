package log_test

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/k3/log"
	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestAvailableCategories(t *testing.T) {
	conf, err := project.NewConfig(project.WithK3())
	if err != nil {
		t.Fatal(err)
	}

	k3r := conf.K3.Runtime
	if k3r == "k3r" || k3r == "" {
		t.Skip("no k3 runtime found")
	}

	cmd := proc.Command(k3r, "--describe-events")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	actual := make(map[string]log.Category)
	reg := regexp.MustCompile(`^ *\d+: *`)
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		name := s.Text()
		c := name
		if len(c) != 4 {
			t.Errorf("Invalid category name %q", c)
		}

		sep := " "
		for s.Scan() {
			if !strings.HasPrefix(s.Text(), "  ") {
				break
			}
			line := strings.TrimSpace(s.Text())

			if line == "--" {
				sep = "|"
				continue
			}
			c = c + sep + reg.ReplaceAllString(line, "")
		}

		// The first separator is a space and should be replaced with a pipe.
		c = strings.Replace(c, " ", "|", 1)

		if _, ok := actual[name]; !ok {
			actual[name] = log.Category(c)
		} else {
			t.Errorf("Category %q already exists", name)
		}
	}

	assert.Equal(t, log.Categories, actual)
}
