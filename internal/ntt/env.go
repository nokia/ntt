package ntt

import (
	"fmt"
	"os"
	"strings"
)

// Getenv retrieves the value of the environment variable named by the key. It
// returns the value, which will be empty if the variable is not present.
//
// If key starts with "NTT" and could not be found, Getenv will also try the key
// with "K3" prefix. For example if key "NTT_IMPORTS" could not be found, Getenv
// would also try key "K3_IMPORTS".
func (s *Suite) Getenv(v string) string {
	if env := os.Getenv(v); env != "" {
		return env
	}

	// This extra lookup with "K3" prefix helps to migrate old bash-scripts to
	// this ntt-package.
	if len(v) >= 3 && v[:3] == "NTT" {
		return s.Getenv("K3" + strings.TrimPrefix(v, "NTT"))
	}

	return ""
}

// expand expands string v trying getenv. Unset environment variables wont get
// substituted.
func (s *Suite) expand(v string) string {
	mapper := func(name string) string {
		val, ok := os.LookupEnv(name)
		if ok {
			return val
		}

		// Don't expand
		return fmt.Sprintf("${%s}", name)
	}

	return os.Expand(v, mapper)
}
