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

	b2, err := format.Source(buf.Bytes())
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
