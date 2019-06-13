package task

import "os"

// Getenv returns the value of environment variable `name` or default value v if
// not set.
func Getenv(name string, v string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return v
}
