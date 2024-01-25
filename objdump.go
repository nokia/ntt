package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/spf13/cobra"
)

var (
	ObjdumpCommand = &cobra.Command{
		Use:   "objdump",
		Short: "Display information from T3XF object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			indent = 0 // TODO(5nord): clean up global variables (sorry).
			offset := 0
			for offset < len(b) {
				n, op, arg := t3xf.Decode(b[offset:])
				printInstruction(offset, b, op, arg)
				fmt.Fprintf(w, "\n")
				offset += n
			}
			w.Flush()
			return nil
		},
	}

	line    = color.New(color.FgWhite, color.Faint)
	literal = color.New()
)

func printInstruction(offset int, b []byte, op opcode.Opcode, arg interface{}) {
	switch op {
	case opcode.SCAN:
		printOffset(offset)
		line.Fprintf(w, "scan")
		indent += 4
	case opcode.BLOCK:
		indent -= 4
		printOffset(offset)
		line.Fprintf(w, "block")
	case opcode.LINE:
		line.Fprintf(w, "          %*s=%d", indent, "", arg)
	case opcode.REF, opcode.IEEE754DP, opcode.NATLONG, opcode.ISTR, opcode.FSTR:
		printOffset(offset)
		printArgument(arg)
	case opcode.UTF8:
		printOffset(offset)
		literal.Fprintf(w, "%q", arg.(string))
	default:
		printOffset(offset)
		fmt.Fprintf(w, "%s", op.String())
		if arg != nil {
			fmt.Fprintf(w, " ")
			printArgument(arg)
		}
	}
}

func printOffset(offset int) {
	line.Fprintf(w, "%08x: %*s", offset, indent, "")
}

func printArgument(arg interface{}) {
	switch arg := arg.(type) {
	case int:
		literal.Fprintf(w, "%d", arg)
	case float64:
		literal.Fprintf(w, "%f", arg)
	case string:
		literal.Fprintf(w, "%s", arg)
	case t3xf.Reference:
		bold.Fprintf(w, "@%08x", arg)
	case *t3xf.String:
		literal.Fprintf(w, "0x%02x (len=%d)", arg.Bytes(), arg.Len())
	default:
		fmt.Fprintf(w, "%+v", arg)
	}
}
