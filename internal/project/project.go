// Package project collects information about test suite organisation by
// implementing various heuristics.
//
// A language server, for example, must always know which TTCN-3 module belongs
// to what test suite. However, there's no standard way of telling where to
// actually look for. Not always do have a manifest file available. Therefore
// we must rely on various heuristics and provide a best effort solution.
package project

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
