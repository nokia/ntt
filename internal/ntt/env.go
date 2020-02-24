package ntt

import (
	"fmt"
	"os"
	"strings"
)

func getenv(s string) string {
	if x := os.Getenv("NTT_" + strings.ToUpper(s)); x != "" {
		return x
	}
	if x := os.Getenv("K3_" + strings.ToUpper(s)); x != "" {
		return x
	}
	return ""
}

// expand expands s trying getenv. Unset environment variables wont get
// substituted.
func expand(s string) string {
	mapper := func(name string) string {
		val, ok := os.LookupEnv(name)
		if ok {
			return val
		}

		// Don't expand
		return fmt.Sprintf("${%s}", name)
	}

	return os.Expand(s, mapper)
}
