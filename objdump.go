package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/spf13/cobra"
)

type t3xfDumper interface {
	printInstruction(int, []byte, opcode.Opcode, interface{})
	acquireRefs(int, int)
}

type t3xfAsmDump struct {
	indent          int
	lineNo          int
	offsetToLineMap map[int]int
}

type t3xfObjDump struct {
	indent int
}

func NewT3xfObjDump() *t3xfObjDump {
	return &t3xfObjDump{indent: 0}
}
func NewT3xfAsmDump() *t3xfAsmDump {
	return &t3xfAsmDump{indent: 0, lineNo: 1, offsetToLineMap: make(map[int]int)}
}

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
			offset := 0
			var dumper t3xfDumper
			if !asT3xfAsm {
				dumper = NewT3xfObjDump()
			} else {
				dumper = NewT3xfAsmDump()
				lineNo := 1
				for offset < len(b) {
					n, _, _ := t3xf.Decode(b[offset:])
					dumper.acquireRefs(offset, lineNo)
					lineNo++
					offset += n
				}
				offset = 0
			}
			for offset < len(b) {
				n, op, arg := t3xf.Decode(b[offset:])
				dumper.printInstruction(offset, b, op, arg)
				fmt.Fprintf(w, "\n")
				offset += n
			}
			w.Flush()
			return nil
		},
	}

	line      = color.New(color.FgWhite, color.Faint)
	literal   = color.New()
	asT3xfAsm bool
)

const INDENT_OFFSET = 4

func init() {
	ObjdumpCommand.Flags().BoolVarP(&asT3xfAsm, "t3xfasm", "a", false, "output as t3xf assembler (suitable as input for 'ntt t3xfasm')")
}

func (asmDump *t3xfAsmDump) acquireRefs(offset int, asmLineNo int) {
	asmDump.offsetToLineMap[offset] = asmLineNo
}

func (asmDump *t3xfAsmDump) printInstruction(offset int, b []byte, op opcode.Opcode, arg interface{}) {
	asmDump.lineNo++
	switch op {
	case opcode.SCAN, opcode.ISCAN:
		line.Fprintf(w, "%*sscan", asmDump.indent, "")
		asmDump.indent += INDENT_OFFSET
	case opcode.BLOCK:
		asmDump.indent -= INDENT_OFFSET
		line.Fprintf(w, "%*sblock", asmDump.indent, "")
	case opcode.LINE:
		line.Fprintf(w, "%*s=%d", asmDump.indent, "", arg)
	case opcode.NATLONG:
		line.Fprintf(w, "%*snatlong ", asmDump.indent, "")
		asmDump.printArgument(arg)
	case opcode.ISTR:
		line.Fprintf(w, "%*sistr ", asmDump.indent, "")
		asmDump.printArgument(arg)
	case opcode.FSTR:
		line.Fprintf(w, "%*sfstr ", asmDump.indent, "")
		asmDump.printArgument(arg)
	case opcode.IEEE754DP:
		line.Fprintf(w, "%*sieee754dp ", asmDump.indent, "")
		asmDump.printArgument(arg)
	case opcode.REF:
		line.Fprintf(w, "%*s", asmDump.indent, "")
		asmDump.printArgument(arg)
	case opcode.UTF8:
		literal.Fprintf(w, "%*sutf8 %q", asmDump.indent, "", arg.(string))
	default:
		fmt.Fprintf(w, "%*s%s", asmDump.indent, "", op.String())
		if arg != nil {
			fmt.Fprintf(w, " ")
			asmDump.printArgument(arg)
		}
	}
}

func (*t3xfObjDump) acquireRefs(int, int) {
}

func (objDump *t3xfObjDump) printInstruction(offset int, b []byte, op opcode.Opcode, arg interface{}) {
	switch op {
	case opcode.SCAN, opcode.ISCAN:
		objDump.printOffset(offset)
		line.Fprintf(w, "scan")
		objDump.indent += INDENT_OFFSET
	case opcode.BLOCK:
		objDump.indent -= INDENT_OFFSET
		objDump.printOffset(offset)
		line.Fprintf(w, "block")
	case opcode.LINE:
		line.Fprintf(w, "          %*s=%d", objDump.indent, "", arg)
	case opcode.REF, opcode.IEEE754DP, opcode.NATLONG, opcode.ISTR, opcode.FSTR:
		objDump.printOffset(offset)
		objDump.printArgument(arg)
	case opcode.UTF8:
		objDump.printOffset(offset)
		literal.Fprintf(w, "%q", arg.(string))
	default:
		objDump.printOffset(offset)
		fmt.Fprintf(w, "%s", op.String())
		if arg != nil {
			fmt.Fprintf(w, " ")
			objDump.printArgument(arg)
		}
	}
}

func (objDump *t3xfObjDump) printOffset(offset int) {
	line.Fprintf(w, "%08x: %*s", offset, objDump.indent, "")
}

func (objDump *t3xfObjDump) printArgument(arg interface{}) {
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

func (asmDump *t3xfAsmDump) printArgument(arg interface{}) {
	switch arg := arg.(type) {
	case int:
		literal.Fprintf(w, "%d", arg)
	case float64:
		literal.Fprintf(w, "%g", arg)
	case string:
		literal.Fprintf(w, "\"%s\"", arg)
	case t3xf.Reference:
		bold.Fprintf(w, "@%d", asmDump.offsetToLineMap[int(arg)])
	case *t3xf.String:
		literal.Fprintf(w, "0x%02x (len=%d)", arg.Bytes(), arg.Len())
	default:
		fmt.Fprintf(w, "%+v", arg)
	}
}
