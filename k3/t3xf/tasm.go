package t3xf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

func Read(r io.Reader) (*File, error) {

	magic := make([]byte, 8)
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if bytes.Compare(magic, []byte("T3XFASM\x00")) != 0 {
		return nil, errors.New("invalid magic number")
	}

	header := fileMap{}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	file := File{
		T3xf:             make([]byte, header.T3xf),
		Text:             make([]byte, header.Text),
		Int32s:           make([]Int32Entry, int(header.Int32s)/binary.Size(Int32Entry{})),
		Names:            make([]NameEntry, int(header.Names)/binary.Size(NameEntry{})),
		Modules:          make([]ModuleEntry, int(header.Modules)/binary.Size(ModuleEntry{})),
		TypeAliass:       make([]TypeAliasEntry, int(header.TypeAliass)/binary.Size(TypeAliasEntry{})),
		RecordTypes:      make([]RecordTypeEntry, int(header.RecordTypes)/binary.Size(RecordTypeEntry{})),
		SetTypes:         make([]SetTypeEntry, int(header.SetTypes)/binary.Size(SetTypeEntry{})),
		RecordOfTypes:    make([]RecordOfTypeEntry, int(header.RecordOfTypes)/binary.Size(RecordOfTypeEntry{})),
		SetOfTypes:       make([]SetOfTypeEntry, int(header.SetOfTypes)/binary.Size(SetOfTypeEntry{})),
		UnionTypes:       make([]UnionTypeEntry, int(header.UnionTypes)/binary.Size(UnionTypeEntry{})),
		EnumeratedTypes:  make([]EnumeratedTypeEntry, int(header.EnumeratedTypes)/binary.Size(EnumeratedTypeEntry{})),
		ArrayTypes:       make([]ArrayTypeEntry, int(header.ArrayTypes)/binary.Size(ArrayTypeEntry{})),
		ClosureTypes:     make([]ClosureTypeEntry, int(header.ClosureTypes)/binary.Size(ClosureTypeEntry{})),
		MessagePortTypes: make([]MessagePortTypeEntry, int(header.MessagePortTypes)/binary.Size(MessagePortTypeEntry{})),
		ComponentTypes:   make([]ComponentTypeEntry, int(header.ComponentTypes)/binary.Size(ComponentTypeEntry{})),
		Consts:           make([]ConstEntry, int(header.Consts)/binary.Size(ConstEntry{})),
		ModulePars:       make([]ModuleParEntry, int(header.ModulePars)/binary.Size(ModuleParEntry{})),
		Templates:        make([]TemplateEntry, int(header.Templates)/binary.Size(TemplateEntry{})),
		Testcases:        make([]TestcaseEntry, int(header.Testcases)/binary.Size(TestcaseEntry{})),
		Functions:        make([]FunctionEntry, int(header.Functions)/binary.Size(FunctionEntry{})),
		ExtFunctions:     make([]ExtFunctionEntry, int(header.ExtFunctions)/binary.Size(ExtFunctionEntry{})),
		Altsteps:         make([]AltstepEntry, int(header.Altsteps)/binary.Size(AltstepEntry{})),
		Blocks:           make([]BlockEntry, int(header.Blocks)/binary.Size(BlockEntry{})),
		Controls:         make([]ControlEntry, int(header.Controls)/binary.Size(ControlEntry{})),
		Strings:          make([]StringEntry, int(header.Strings)/binary.Size(StringEntry{})),
		Collections:      make([]CollectionEntry, int(header.Collections)/binary.Size(CollectionEntry{})),
		Data:             make([]byte, header.Data),
	}

	var err error
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.T3xf))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Text))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Int32s))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Names))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Modules))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.TypeAliass))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.RecordTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.SetTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.RecordOfTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.SetOfTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.UnionTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.EnumeratedTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.ArrayTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.ClosureTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.MessagePortTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.ComponentTypes))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Consts))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.ModulePars))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Templates))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Testcases))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Functions))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.ExtFunctions))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Altsteps))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Blocks))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Controls))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Strings))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Collections))
	err = errors.Join(err, binary.Read(r, binary.LittleEndian, &file.Data))

	if err != nil {
		return nil, err
	}

	return &file, nil
}

type File struct {
	T3xf             []byte
	Text             []byte
	Int32s           []Int32Entry
	Names            []NameEntry
	Modules          []ModuleEntry
	TypeAliass       []TypeAliasEntry
	RecordTypes      []RecordTypeEntry
	SetTypes         []SetTypeEntry
	RecordOfTypes    []RecordOfTypeEntry
	SetOfTypes       []SetOfTypeEntry
	UnionTypes       []UnionTypeEntry
	EnumeratedTypes  []EnumeratedTypeEntry
	ArrayTypes       []ArrayTypeEntry
	ClosureTypes     []ClosureTypeEntry
	MessagePortTypes []MessagePortTypeEntry
	ComponentTypes   []ComponentTypeEntry
	Consts           []ConstEntry
	ModulePars       []ModuleParEntry
	Templates        []TemplateEntry
	Testcases        []TestcaseEntry
	Functions        []FunctionEntry
	ExtFunctions     []ExtFunctionEntry
	Altsteps         []AltstepEntry
	Blocks           []BlockEntry
	Controls         []ControlEntry
	Strings          []StringEntry
	Collections      []CollectionEntry
	Data             []byte
}

type fileMap struct {
	T3xf             uint32
	Text             uint32
	Int32s           uint32
	Names            uint32
	Modules          uint32
	TypeAliass       uint32
	RecordTypes      uint32
	SetTypes         uint32
	RecordOfTypes    uint32
	SetOfTypes       uint32
	UnionTypes       uint32
	EnumeratedTypes  uint32
	ArrayTypes       uint32
	ClosureTypes     uint32
	MessagePortTypes uint32
	ComponentTypes   uint32
	Consts           uint32
	ModulePars       uint32
	Templates        uint32
	Testcases        uint32
	Functions        uint32
	ExtFunctions     uint32
	Altsteps         uint32
	Blocks           uint32
	Controls         uint32
	Strings          uint32
	Collections      uint32
	Data             uint32
}

type Int32Entry struct {
	AssetOffset uint32
	Value       int32
}
type NameEntry struct {
	AssetOffset uint32
	StringId    uint32
}

type ModuleEntry struct {
	ModuleStringId    uint32
	SourceStringId    uint32
	BlockId           uint32
	HasControlSection uint16
}

type TypeAliasEntry struct {
	AssetOffset        uint32
	ModuleId           uint16
	NameStringId       uint32
	RootTypeFileOffset uint32
	WithBlockId        uint32
	ConstraintBlockId  uint32
	LineNumber         uint32
}

type RecordTypeEntry struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	LineNumber        uint32
}

type SetTypeEntry struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	ConstraintBlockId uint32
	WithBlockId       uint32
	LineNumber        uint32
}

type RecordOfTypeEntry struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	ConstraintBlockId uint32
	WithBlockId       uint32
	LineNumber        uint32
}

type SetOfTypeEntry struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	ElementTypeFileOffset uint32
	ConstraintBlockId     uint32
	WithBlockId           uint32
	LineNumber            uint32
}

type UnionTypeEntry struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	LineNumber        uint32
}

type EnumeratedTypeEntry struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	LineNumber        uint32
}

type ArrayTypeEntry struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	ElementTypeFileOffset uint32
	DimensionBlockId      uint32
	WithBlockId           uint32
	LineNumber            uint32
}

type ClosureTypeEntry struct {
	NameStringId uint32
	ModuleId     uint16
	AssetOffset  uint32
	WithBlockId  uint32
	LineNumber   uint32
}

type MessagePortTypeEntry struct {
	AssetOffset                  uint32
	ModuleId                     uint16
	NameStringId                 uint32
	InDefinitionBlockId          uint32
	OutDefinitionBlockId         uint32
	MapParamsDefinitionBlockId   uint32
	UnmapParamsDefinitionBlockId uint32
	AddressType                  uint32
	WithBlockId                  uint32
	LineNumber                   uint32
}

type ComponentTypeEntry struct {
	AssetOffset                uint32
	ModuleId                   uint16
	NameStringId               uint32
	ExtendingComponentsBlockId uint32
	InitialisationBlockId      uint32
	WithBlockId                uint32
	LineNumber                 uint32
}

type ConstEntry struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	TypeFileOffset        uint32
	InitialisationBlockId uint32
	WithBlockId           uint32
	LineNumber            uint32
}

type ModuleParEntry struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	TypeFileOffset        uint32
	InitialisationBlockId uint32
	WithBlockId           uint32
	LineNumber            uint32
}

type TemplateEntry struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	TypeFileOffset        uint32
	ParamsBlockId         uint32
	InitialisationBlockId uint32
	LineNumber            uint32
}
type TestcaseEntry struct {
	NameStringId     uint32
	ModuleId         uint16
	RunsOnFileOffset uint32
	SystemFileOffset uint32
	BodyBlockId      uint32
	ParamsBlockId    uint32
	AssetOffset      uint32
	LineNumber       uint32
}

type FunctionEntry struct {
	NameStringId         uint32
	ModuleId             uint16
	RunsOnFileOffset     uint32
	ReturnTypeFileOffset uint32
	ReturnTypeFlags      uint16
	BodyBlockId          uint32
	ParamsBlockId        uint32
	AssetOffset          uint32
	LineNumber           uint32
}

type ExtFunctionEntry struct {
	NameStringId         uint32
	ModuleId             uint16
	ReturnTypeFileOffset uint32
	ReturnTypeFlags      uint16
	ParamsBlockId        uint32
	AssetOffset          uint32
	WithBlockId          uint32
	LineNumber           uint32
}

type AltstepEntry struct {
	NameStringId     uint32
	ModuleId         uint16
	RunsOnFileOffset uint32
	BodyBlockId      uint32
	ParamsBlockId    uint32
	AssetOffset      uint32
	LineNumber       uint32
}

type BlockEntry struct {
	AssetOffset uint32
	TextOffset  uint32
	Length      uint32
}

type ControlEntry struct {
	AssetOffset uint32
	ModuleId    uint16
	BlockId     uint32
	Length      uint32
	LineNumber  uint32
}

type StringEntry struct {
	DataOffset uint32
	Length     uint32
}

type CollectionEntry struct {
	DataOffset uint32
	Length     uint32
}
