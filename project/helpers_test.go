package project

// AcquireExecutables depending on the provided presets and on the availability
// inside the ttcn-3 code, a list of executable ttcn-3 entities (i.e. testcases,
// control parts) is returned
func AcquireExecutables(gc *Parameters, files []string, presets []string) []TestConfig {
	return acquireExecutables(gc, files, presets)
}
