package project

// Build struct provides information about build directory organisation.
// Usually the build system write this information into the file
// ttcn3_suites.json
type Build struct {
	SourceDir string  `json:"source_dir"`
	BinaryDir string  `json:"binary_dir"`
	Suites    []Suite `json:"suites"`
}

// Suite struct describes where to find the manifest and the source directory
// in a test suite.
type Suite struct {
	RootDir   string `json:"root_dir"`
	SourceDir string `json:"source_dir"`
}
