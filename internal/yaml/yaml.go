// Package yaml provides uniform interface for parsing YAML files.
package yaml

// I tried various YAML libraries, but they all had their own "quirks".
//
// * gopkg.in/yaml.v2: Unmarshalling embedded types is tricky, error messages
// are okay.
//
// * gopkg.in/yaml.v3: Not sure if this is still alive. Silently ignores
// unknown fields.
//
// * "github.com/goccy/go-yaml": Has nice error messages, re-uses the json tag,
// but also ignores unknown fields. Might be a YAML thing?
import (
	//"gopkg.in/yaml.v3"
	"fmt"
	"strconv"
	"time"

	"github.com/goccy/go-yaml"
)

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	f, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return fmt.Errorf("%s: %w", string(text), err)
	}

	*d = Duration{time.Duration(f * float64(time.Second))}
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%g", d.Seconds())), nil
}

func Unmarshal(data []byte, v interface{}) error {
	return yaml.UnmarshalWithOptions(data, v, yaml.Strict())
}

func Marshal(v interface{}) ([]byte, error) {
	return yaml.MarshalWithOptions(v, yaml.UseLiteralStyleIfMultiline(true))
}

func MarshalJSON(v interface{}) ([]byte, error) {
	return yaml.MarshalWithOptions(v, yaml.JSON())
}
