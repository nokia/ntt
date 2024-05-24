// This program helps to modify opcodes.yml programmatically.

package main

import (
	"os"

	"github.com/nokia/ntt/internal/yaml"
	"github.com/nokia/ntt/k3/t3xf/opcode"
)

var disclaimer = `# This file contains the T3XF instructions used by K3 runtime and MTC compiler.
# It contains the opcode, stack operations, execution context and a
# description, if available.
#
# When you update this file please run 'go generate' to update the implemtation.
# It is recommended to also run 'go test ./...' from inside the t3xf directory.

`

var opcodes_file = "./opcodes.yml"

func main() {
	// Read opcode descriptions from ../opcodes.yml
	opcodes := make(map[string]*opcode.Descriptor)
	b, err := os.ReadFile(opcodes_file)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(b, &opcodes); err != nil {
		panic(err)
	}

	removeContext(opcodes)

	// Write opcode descriptions to ../opcodes.yml
	b, err = yaml.Marshal(opcodes)
	if err != nil {
		panic(err)
	}

	b = append([]byte(disclaimer), b...)
	if err := os.WriteFile(opcodes_file, b, 0644); err != nil {
		panic(err)
	}

}

// The context field contains the C++ Environment classes in with the opcode is
// used. It is much more useful to have the kind of objects the instructions
// operation on instead.
func removeContext(opcodes map[string]*opcode.Descriptor) {
	for _, v := range opcodes {
		v.Context = nil
	}
}
