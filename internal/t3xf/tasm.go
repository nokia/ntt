package t3xf

type tasmHeader struct {
	T3xfSection          uint32
	TextSection          uint32
	Int32Table           uint32
	NameTable            uint32
	ModuleTable          uint32
	TypeAliasTable       uint32
	RecordTypeTable      uint32
	SetTypeTable         uint32
	RecordOfTypeTable    uint32
	SetOfTypeTable       uint32
	UnionTypeTable       uint32
	EnumeratedTypeTable  uint32
	ArrayTypeTable       uint32
	ClosureTypeTable     uint32
	MessagePortTypeTable uint32
	ComponentTypeTable   uint32
	ConstTable           uint32
	ModuleParTable       uint32
	TemplateTable        uint32
	TestcaseTable        uint32
	FunctionTable        uint32
	ExtFunctionTable     uint32
	AltstepTable         uint32
	BlockTable           uint32
	ControlTable         uint32
	StringTable          uint32
	CollectionTable      uint32
	DataSection          uint32
}

type Sections struct {
	T3XF []byte
	text []byte
	data []byte
}

type Tables struct {
	altsteps         []tasmAltstep
	arrayTypes       []tasmArray
	blocks           []tasmBlock
	closureTypes     []tasmClosure
	collections      []tasmCollection
	componentTypes   []tasmComponent
	consts           []tasmConst
	controls         []tasmControl
	enumeratedTypes  []tasmEnum
	extFunctions     []tasmExtFunction
	functions        []tasmFunction
	ints             []tasmInt32
	messagePortTypes []tasmPort
	modulePars       []tasmConst
	modules          []tasmModule
	names            []tasmName
	recordOfTypes    []tasmRecordOf
	recordTypes      []tasmRecord
	setOfTypes       []tasmSetOf
	setTypes         []tasmSet
	strings          []tasmString
	templates        []tasmTemplate
	testcases        []tasmTestcase
	typeAliases      []tasmSubType
	unionTypes       []tasmUnion
}

type tasmModule struct {
	ModuleStringId uint32
	SourceStringId uint32
	BlockId        uint32
	HasControl     uint16
}

type tasmString struct {
	DataOffset uint32
	Length     uint32
}

type tasmCollection struct {
	DataOffset uint32
	Length     uint32
}

type tasmInt32 struct {
	AssetOffset uint32
	Value       int32
}

type tasmName struct {
	AssetOffset uint32
	StringId    uint32
}

type tasmBlock struct {
	AssetOffset uint32
	TextOffset  uint32
	Length      uint32
}

type tasmControl struct {
	AssetOffset uint32
	ModuleId    uint16
	BlockId     uint32
	Length      uint32
	Line        uint32
}

type tasmTestcase struct {
	NameStringId     uint32
	ModuleId         uint16
	RunsOnFileOffset uint32
	SystemFileOffset uint32
	BodyBlockId      uint32
	ParamsBlockId    uint32
	AssetOffset      uint32
	Line             uint32
}

type tasmFunction struct {
	NameStringId         uint32
	ModuleId             uint16
	RunsOnFileOffset     uint32
	ReturnTypeFileOffset uint32
	ReturnTypeFlags      uint16
	BodyBlockId          uint32
	ParamsBlockId        uint32
	AssetOffset          uint32
	Line                 uint32
}

type tasmExtFunction struct {
	NameStringId         uint32
	ModuleId             uint16
	ReturnTypeFileOffset uint32
	ReturnTypeFlags      uint16
	ParamsBlockId        uint32
	AssetOffset          uint32
	WithBlockId          uint32
	Line                 uint32
}

type tasmAltstep struct {
	NameStringId     uint32
	ModuleId         uint16
	RunsOnFileOffset uint32
	BodyBlockId      uint32
	ParamsBlockId    uint32
	AssetOffset      uint32
	Line             uint32
}

type tasmClosure struct {
	NameStringId uint32
	ModuleId     uint16
	AssetOffset  uint32
	WithBlockId  uint32
	Line         uint32
}

type tasmSubType struct {
	AssetOffset        uint32
	ModuleId           uint16
	NameStringId       uint32
	RootTypeFileOffset uint32
	WithBlockId        uint32
	ConstraintBlockId  uint32
	Line               uint32
}

type tasmRecord struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	Line              uint32
}

type tasmSet struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	Line              uint32
}

type tasmRecordOf struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	ConstraintBlockId uint32
	WithBlockId       uint32
	Line              uint32
}

type tasmSetOf struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	ElementTypeFileOffset uint32
	ConstraintBlockId     uint32
	WithBlockId           uint32
	Line                  uint32
}

type tasmUnion struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	Line              uint32
}

type tasmEnum struct {
	AssetOffset       uint32
	ModuleId          uint16
	NameStringId      uint32
	DefinitionBlockId uint32
	WithBlockId       uint32
	ConstraintBlockId uint32
	Line              uint32
}

type tasmArray struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	ElementTypeFileOffset uint32
	DimensionBlockId      uint32
	WithBlockId           uint32
	Line                  uint32
}

type tasmPort struct {
	AssetOffset                  uint32
	ModuleId                     uint16
	NameStringId                 uint32
	InDefinitionBlockId          uint32
	OutDefinitionBlockId         uint32
	MapParamsDefinitionBlockId   uint32
	UnmapParamsDefinitionBlockId uint32
	Addresstype                  uint32
	WithBlockId                  uint32
	Line                         uint32
}

type tasmComponent struct {
	AssetOffset                uint32
	ModuleId                   uint16
	NameStringId               uint32
	ExtendingComponentsBlockId uint32
	InitialisationBlockId      uint32
	WithBlockId                uint32
	Line                       uint32
}

type tasmConst struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	TypeFileOffset        uint32
	InitialisationBlockId uint32
	WithBlockId           uint32
	Line                  uint32
}

type tasmTemplate struct {
	AssetOffset           uint32
	ModuleId              uint16
	NameStringId          uint32
	TypeFileOffset        uint32
	ParamsBlockId         uint32
	InitialisationBlockId uint32
	Line                  uint32
}
