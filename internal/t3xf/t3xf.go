// Package t3xf provides routines for decoding and loading t3xf encoded files.
//
// t3xf (TTCN-3 Executable Format) is a binary representation of input TTCN-3
// source text, suitable for execution by a native application, its main purpose
// expected to be the execution of test cases against a system under test.
//
// t3xf can be thought of as one possible binary dump of the abstract syntax
// tree created by a TTCN-3 compiler after all semantic checking has been
// performed. t3xf is conceived as a file format. No provision is made for
// streaming or for editing. Transmission of whole files is not only possible
// but will be required in systems that support testing distributed across a
// number of host machines.
//
// tasm (T3xf Assembly) is an extension to t3xf allowing objects to be loaded
// lazily and hence reducing the time for loading t3xf files.
package t3xf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
)

// File represents a t3xf file.
type File struct {
	Sections Sections
	Tables   Tables

	version int              // T3XF Version
	tasm    bool             // true, if file is in TASM format
	width   int              // Instruction width in bytes
	order   binary.ByteOrder // Endianess of file
}

// ReadFile reads a file from path and returns a File struct. An error is returned if
// the file could not be read or has an unknown or invalid format.
func ReadFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	size := 0
	if fi, err := f.Stat(); err == nil {
		// Check if size fits in int
		if fi.Size() <= int64(int(^uint(0)>>1)) {
			size = int(fi.Size())
		}
	}

	return read(f, size)
}

func read(r io.Reader, size int) (*File, error) {

	// If r does not implement Peek(), wrap it into a bufio.Reader
	if _, ok := r.(peeker); !ok {
		r = bufio.NewReader(r)
	}

	er := binaryReader{r: r, size: size}

	switch {
	case isT3XF(r.(peeker)):
		b, err := ioutil.ReadAll(r)
		return &File{
			Sections: Sections{T3XF: b},
			version:  2,
			width:    4,
			order:    binary.LittleEndian,
		}, err

	case isTASM(r.(peeker)):
		magic := make([]byte, 8)
		er.read(&magic)

	default:
		return nil, ErrUnknownFormat
	}

	var h tasmHeader
	er.read(&h)

	count := func(n uint32, v interface{}) int {
		return int(n / uint32(binary.Size(v)))
	}

	f := File{
		tasm:    true,
		version: 1,
		width:   4,
		order:   binary.LittleEndian,

		Sections: Sections{
			T3XF: make([]byte, h.T3xfSection),
			text: make([]byte, h.TextSection),
			data: make([]byte, h.DataSection),
		},
		Tables: Tables{
			ints:             make([]tasmInt32, count(h.Int32Table, tasmInt32{})),
			names:            make([]tasmName, count(h.NameTable, tasmName{})),
			modules:          make([]tasmModule, count(h.ModuleTable, tasmModule{})),
			typeAliases:      make([]tasmSubType, count(h.TypeAliasTable, tasmSubType{})),
			recordTypes:      make([]tasmRecord, count(h.RecordTypeTable, tasmRecord{})),
			setTypes:         make([]tasmSet, count(h.SetTypeTable, tasmSet{})),
			recordOfTypes:    make([]tasmRecordOf, count(h.RecordOfTypeTable, tasmRecordOf{})),
			setOfTypes:       make([]tasmSetOf, count(h.SetOfTypeTable, tasmSetOf{})),
			unionTypes:       make([]tasmUnion, count(h.UnionTypeTable, tasmUnion{})),
			enumeratedTypes:  make([]tasmEnum, count(h.EnumeratedTypeTable, tasmEnum{})),
			arrayTypes:       make([]tasmArray, count(h.ArrayTypeTable, tasmArray{})),
			closureTypes:     make([]tasmClosure, count(h.ClosureTypeTable, tasmClosure{})),
			messagePortTypes: make([]tasmPort, count(h.MessagePortTypeTable, tasmPort{})),
			componentTypes:   make([]tasmComponent, count(h.ComponentTypeTable, tasmComponent{})),
			consts:           make([]tasmConst, count(h.ConstTable, tasmConst{})),
			modulePars:       make([]tasmConst, count(h.ModuleParTable, tasmConst{})),
			templates:        make([]tasmTemplate, count(h.TemplateTable, tasmTemplate{})),
			testcases:        make([]tasmTestcase, count(h.TestcaseTable, tasmTestcase{})),
			functions:        make([]tasmFunction, count(h.FunctionTable, tasmFunction{})),
			extFunctions:     make([]tasmExtFunction, count(h.ExtFunctionTable, tasmExtFunction{})),
			altsteps:         make([]tasmAltstep, count(h.AltstepTable, tasmAltstep{})),
			blocks:           make([]tasmBlock, count(h.BlockTable, tasmBlock{})),
			controls:         make([]tasmControl, count(h.ControlTable, tasmControl{})),
			strings:          make([]tasmString, count(h.StringTable, tasmString{})),
			collections:      make([]tasmCollection, count(h.CollectionTable, tasmCollection{})),
		},
	}

	// Note, below fields must be read in this order.
	er.read(&f.Sections.T3XF)
	er.read(&f.Sections.text)
	er.read(&f.Tables.ints)
	er.read(&f.Tables.names)
	er.read(&f.Tables.modules)
	er.read(&f.Tables.typeAliases)
	er.read(&f.Tables.recordTypes)
	er.read(&f.Tables.setTypes)
	er.read(&f.Tables.recordOfTypes)
	er.read(&f.Tables.setOfTypes)
	er.read(&f.Tables.unionTypes)
	er.read(&f.Tables.enumeratedTypes)
	er.read(&f.Tables.arrayTypes)
	er.read(&f.Tables.closureTypes)
	er.read(&f.Tables.messagePortTypes)
	er.read(&f.Tables.componentTypes)
	er.read(&f.Tables.consts)
	er.read(&f.Tables.modulePars)
	er.read(&f.Tables.templates)
	er.read(&f.Tables.testcases)
	er.read(&f.Tables.functions)
	er.read(&f.Tables.extFunctions)
	er.read(&f.Tables.altsteps)
	er.read(&f.Tables.blocks)
	er.read(&f.Tables.controls)
	er.read(&f.Tables.strings)
	er.read(&f.Tables.collections)
	er.read(&f.Sections.data)

	if er.err != nil {
		return nil, er.err
	}

	return &f, nil
}

// Info returns a string describing File format
func (f *File) Info() string {
	switch {
	case f.tasm && f.width == 4 && f.order == binary.LittleEndian:
		return "TASM 32-bit, little-endian, version 2"
	case !f.tasm && f.width == 4 && f.order == binary.LittleEndian:
		return "T3XF 32-bit, little-endian, version 2"
	default:
		panic("unexpected file format")
	}
}

// IsTASM returns true if file uses TASM format.
func (f *File) IsTASM() bool { return f.tasm }

// T3XF version.
func (f *File) Version() int { return f.version }

// Instruction width in bytes.
func (f *File) Width() int { return f.width }

// Instruction byte order.
func (f *File) Order() binary.ByteOrder { return f.order }

// peeker interface provides a Peek method for reading bytes without consuming.
type peeker interface {
	Peek(n int) ([]byte, error)
}

// isT3XF returns true if r begins with the T3XF magic string
func isT3XF(r peeker) bool {

	t3xfMagic := []byte{
		0x03, 0x00, 0x00, 0x13, // NOP
		0x43, 0x01, 0x00, 0x00, // NATLONG
		0x02, 0x00, 0x00, 0x00, // 2
		0x33, 0x00, 0x00, 0x00, // VERSION
	}

	b, err := r.Peek(len(t3xfMagic))
	if err != nil {
		return false
	}
	return bytes.Equal(b, t3xfMagic)
}

// isTASM returns true if r begins with the TASM magic string
func isTASM(r peeker) bool {

	tasmMagic := []byte{
		0x54, 0x33, 0x58, 0x46, 0x41, 0x53, 0x4d, 0x00, // T3XFASM\0
	}

	b, err := r.Peek(len(tasmMagic))
	if err != nil {
		return false
	}
	return bytes.Equal(b, tasmMagic)
}

// binaryReader provides common read methods and convenient error-checking.
//
// binaryReader uses a neat technique to avoid repetitive error handling code.
// For details have a look at: https://blog.golang.org/errors-are-values
type binaryReader struct {
	r    io.Reader
	size int
	err  error
}

func (er *binaryReader) read(data interface{}) {
	if er.err != nil {
		return
	}
	er.err = binary.Read(er.r, binary.LittleEndian, data)
}
