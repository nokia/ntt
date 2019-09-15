package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/nokia/ntt/internal/t3xf"
	"github.com/nokia/ntt/internal/t3xf/opcode"
)

type printer struct {
	w    io.Writer  // A writer where we want to write to
	name string     // Name of file.
	file *t3xf.File // T3XF file object

	addrMap map[int]int
	indent  int
	pos     int

	printAddrs         bool
	printRaw           bool
	printLines         bool
	printLiteralInstrs bool
}

func NewPrinter(path string, w io.Writer) (*printer, error) {
	file, err := t3xf.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &printer{
		name: path,
		file: file,
		w:    w,
	}, nil
}

func (p *printer) Info() {
	fmt.Fprintf(p.w, "%s: %s\n", p.name, p.file.Info())
}

func (p *printer) Print() error {
	if p.file.IsTASM() {
		fmt.Fprintf(p.w, "%s: Warning: limiting output for TASM files.\n", p.name)
	}

	s := t3xf.NewScanner(p.file.Sections.T3XF)

	// Build line map
	if p.printLines {
		p.addrMap = make(map[int]int, len(p.file.Sections.T3XF)/4)
		pos := 1
		for s.Scan() {
			p.addrMap[s.Offset()] = pos
			pos++
		}
		s.Reset()
	}

	for s.Scan() {
		switch op := s.Opcode(); op {
		case opcode.SCAN, opcode.MARK:
			p.printInstr(s, op.String())
			p.indent++

		case opcode.BLOCK, opcode.COLLECT, opcode.VLIST:
			p.indent--
			p.printInstr(s, op.String())

		case opcode.LINE:
			p.printInstr(s, "="+strconv.Itoa(s.Arg()))

		case opcode.REF:
			p.printInstr(s, "L"+strconv.Itoa(p.lookupAddr(s.Arg())))

		case opcode.FROZEN_REF:
			p.printInstr(s, "R"+strconv.Itoa(p.lookupAddr(s.Arg())))

		case opcode.GOTO:
			p.printInstr(s, "@"+strconv.Itoa(p.lookupAddr(s.Arg())))

		case opcode.IDEF, opcode.IGET, opcode.IFIELD:
			p.printInstr(s, op.String()+" "+strconv.Itoa(s.Arg()))

		case opcode.ISTR,
			opcode.FSTR,
			opcode.OCTETS,
			opcode.UTF8,
			opcode.NAME,
			opcode.NIBBLES,
			opcode.BITS,
			opcode.IEEE754DP,
			opcode.NATLONG:
			if p.printLiteralInstrs {
				p.printInstr(s, op.String())
			}
			p.printValue(s)

		default:
			p.printInstr(s, op.String())
		}
		p.pos++
	}

	if s.Err() != nil {
		return errors.New(fmt.Sprintf("%s: %s", p.name, s.Err().Error()))
	}

	return nil
}

func (p *printer) lookupAddr(addr int) int {
	if p.addrMap != nil {
		if i, ok := p.addrMap[addr]; ok {
			return i
		}
	}
	return addr
}

func (p *printer) printInstr(s *t3xf.Scanner, val string) {
	if p.printAddrs {
		fmt.Fprintf(p.w, "%8d: ", s.Offset())
	}

	if p.printRaw {
		fmt.Fprintf(p.w, "% x ", s.Raw()[0:4])
	}

	fmt.Fprintf(p.w, "%*s%s\n", p.indent*2, "", val)
}

func (p *printer) printValue(s *t3xf.Scanner) {
	var addr string
	if p.printAddrs {
		addr = fmt.Sprintf("% 8d: ", s.Offset())
	}

	if p.printRaw {
		b := s.Raw()
		for i := 4; i < len(b); i += 4 {
			fmt.Fprintf(p.w, "%*s% x", len(addr), "", b[i:i+4])
			if i+4 < len(b) {
				fmt.Fprintln(p.w)
			} else {
				fmt.Fprint(p.w, " ")
			}
		}
	}
	fmt.Fprintf(p.w, "%*s", p.indent*2, "")

	switch s.Opcode() {
	case opcode.ISTR, opcode.FSTR:
		fmt.Fprintf(p.w, "%s\n", s.Bytes())
	case opcode.UTF8:
		fmt.Fprintf(p.w, "\"%s\"\n", s.Bytes())
	case opcode.NAME:
		fmt.Fprintf(p.w, "'%s'\n", s.Bytes())
	case opcode.OCTETS:
		fmt.Fprintf(p.w, "'%x'O\n", s.Bytes())
	case opcode.NIBBLES:
		fmt.Fprintf(p.w, "'%x'H\n", s.Bytes())
	case opcode.BITS:
		fmt.Fprintf(p.w, "'%b'B\n", s.Bytes())

	case opcode.IEEE754DP:
		fmt.Fprintf(p.w, "%f\n", s.Float64())
	case opcode.NATLONG:
		fmt.Fprintf(p.w, "%d\n", s.Arg())
	default:
		fmt.Fprintf(p.w, "\n")
	}
}
