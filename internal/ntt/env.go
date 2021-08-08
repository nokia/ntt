package ntt

// Environ returns a copy of strings representing the environment, in the form "key=value".
func (suite *Suite) Environ() ([]string, error) {
	return suite.p.Environ()
}

// Expand expands string v using Suite.Getenv
func (suite *Suite) Expand(v string) (string, error) {
	return suite.p.Expand(v)
}

func (suite *Suite) Getenv(v string) (string, error) {
	return suite.p.Getenv(v)
}
