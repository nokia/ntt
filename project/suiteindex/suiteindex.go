package suiteindex

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

const Name = "ttcn3_suites.json"

type Config struct {
	SourceDir string  `json:"source_dir"`
	BinaryDir string  `json:"binary_dir"`
	Suites    []Suite `json:"suites"`
}

type Suite struct {
	RootDir   string `json:"root_dir"`
	SourceDir string `json:"source_dir"`
}

func ReadFile(file string) (Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	c := Config{}

	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}

	base := filepath.Dir(file)
	for i := range c.Suites {
		if c.Suites[i].RootDir != "" {
			c.Suites[i].RootDir = fix(base, c.Suites[i].RootDir)
		}
		if c.Suites[i].SourceDir != "" {
			c.Suites[i].SourceDir = fix(base, c.Suites[i].SourceDir)
		}
	}
	return c, nil
}

func fix(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
