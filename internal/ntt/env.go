package ntt

import (
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
