package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/nokia/ntt/internal/yaml"
	"github.com/nokia/ntt/k3/t3xf/opcode"
)

var funcs = template.FuncMap{
	"uc":  strings.ToUpper,
	"now": time.Now,
	"encode": func(op int) string {
		return fmt.Sprintf("0x%04x", op*16+3)
	},
	"comment": func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}

		return "// " + strings.ReplaceAll(s, "\n", "\n// ")
	},
	"quote": func(s string) string {
		return fmt.Sprintf("%q", s)
	},
	"join": func(a []string, sep string) string {
		return strings.Join(a, sep)
	},
}

func main() {
	opcodes := make(map[string]*opcode.Descriptor)

	b, err := os.ReadFile("opcodes.yml")
	if err != nil {
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(b, &opcodes); err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("opcode.go.tmpl").Funcs(funcs).ParseFiles("opcode.go.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, opcodes); err != nil {
		log.Fatal(err)
	}
	b = buf.Bytes()

	b2, err := format.Source(b)
	if err == nil {
		b = b2
	}
	// Always write to file to help debugging.
	if err := os.WriteFile("opcode_gen.go", b, 0644); err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatalf("%s: error: %s\n", "opcode_gen.go", err.Error())
	}

}
