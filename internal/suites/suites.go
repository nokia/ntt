// Package suites help finding TTCN-3 suites by implementing various heuristics
// about test suite organization.
package suites

// Build struct provides information about build directory organisation.
// Usually the build system write this information into the file
// ttcn3_suites.json
type Build struct {
	SourceDir string `json:"source_dir"`
	BinaryDir string `json:"binary_dir"`
	Suites    []Suite
}

// Suite struct describes where to find the manifest and the source directory
// in a test suite.
type Suite struct {
	RootDir   string `json:"root_dir"`
	SourceDir string `json:"source_dir"`
}
