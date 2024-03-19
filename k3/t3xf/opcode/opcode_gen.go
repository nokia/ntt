// Package opcode defines the opcodes used in T3XF.
package opcode

const (
	REF        Opcode = 0
	GOTO              = 1
	LINE              = 2
	OPERATION         = 3
	FROZEN_REF        = 4

	ACTION           = 0x1003
	ACTIVATE         = 0x0d03
	ADD              = 0x0413
	ADDRESS          = 0x0ea3
	ALIVE            = 0x1103
	ALIVE1           = 0x1113
	ALIVEA           = 0x1123
	ALLFROM          = 0x06a3
	ALLFROMP         = 0x06b3
	ALT              = 0x0383
	ALTND            = 0x2023
	ALTSTEP          = 0x0803
	ALTSTEPB         = 0x0813
	ALTSTEPBW        = 0x0833
	ALTSTEPW         = 0x0823
	AND              = 0x0483
	ANY              = 0x0603
	ANYN             = 0x0613
	APPLY            = 0x0193
	ARRAY            = 0x0ee3
	ASSIGN           = 0x0583
	ASSIGND          = 0x0593
	AT_DEFAULT       = 0x2013
	BIT2HEX          = 0x18c3
	BIT2INT          = 0x18b3
	BIT2OCT          = 0x18d3
	BIT2STR          = 0x18e3
	BITS             = 0x0113
	BITSTRING        = 0x0e03
	BLOCK            = 0x0083
	BOOLEAN          = 0x0e13
	BREAK            = 0x03c3
	CAT              = 0x0563
	CHAR2INT         = 0x1883
	CHAR2OCT         = 0x1893
	CHARSTRING       = 0x0e23
	CHARSTRINGU      = 0x0e33
	CHECK            = 0x14a3
	CHECK1           = 0x14b3
	CHECKSTATE       = 0x1783
	CHECKSTATEAL     = 0x1793
	CHECKSTATEAN     = 0x17a3
	CLEAR            = 0x1383
	CLEARA           = 0x1393
	CLOSURE          = 0x11f3
	CLOSURETYPE      = 0x0e93
	COLLECT          = 0x0093
	COMPLEMENT       = 0x0623
	COMPONENT        = 0x0f23
	COMPONENTX       = 0x0f33
	CONNECT          = 0x1083
	CONST            = 0x0843
	CONSTW           = 0x0853
	CONTINUE         = 0x03d3
	CONTROL          = 0x0863
	CREATE           = 0x1183
	CREATEA          = 0x1193
	CREATEAN         = 0x11b3
	CREATEN          = 0x11a3
	DEACTIVATE       = 0x0d23
	DEACTIVATEA      = 0x0d33
	DECVALUE         = 0x1a83
	DEF              = 0x05b3
	DEFAULT          = 0x0ec3
	DISCONNECT       = 0x1093
	DISCONNECTA      = 0x10a3
	DISCONNECTAA     = 0x10b3
	DIV              = 0x0453
	DONE             = 0x1143
	DONE1            = 0x1153
	DONEA            = 0x1163
	DOWHILE          = 0x0363
	DROP             = 0x00a3
	ELSE             = 0x03a3
	ENCODE           = 0x0c43
	ENCODEO          = 0x0c63
	ENCVALUE         = 0x1a73
	ENUM2INT         = 0x1a53
	ENUMERATED       = 0x0f03
	EQ               = 0x0503
	ERROR            = 0x02e3
	EXEC             = 0x0373
	EXECUTE          = 0x1063
	EXECUTED         = 0x11d3
	EXECUTEL         = 0x1073
	EXECUTELD        = 0x11c3
	EXTENSION        = 0x0c83
	EXTENSIONO       = 0x0ca3
	FAIL             = 0x02d3
	FALSE            = 0x0283
	FIELD            = 0x0723
	FIELDO           = 0x0733
	FLOAT            = 0x0e43
	FLOAT2INT        = 0x1873
	FOR              = 0x0343
	FROM             = 0x06c3
	FSTR             = 0x0173
	FUNCTION         = 0x0903
	FUNCTIONB        = 0x0913
	FUNCTIONV        = 0x0943
	FUNCTIONVB       = 0x0953
	FUNCTIONX        = 0x0983
	FUNCTIONXV       = 0x09a3
	FUNCTIONXVW      = 0x09b3
	FUNCTIONXW       = 0x0993
	GE               = 0x0513
	GET              = 0x0573
	GETVERDICT       = 0x1043
	GT               = 0x0523
	HALT             = 0x13a3
	HALTA            = 0x13b3
	HEX2BIT          = 0x1903
	HEX2INT          = 0x18f3
	HEX2OCT          = 0x1913
	HEX2STR          = 0x1923
	HEXSTRING        = 0x0e53
	IDEF             = 0x05c3
	IEEE754DP        = 0x0153
	IF               = 0x0303
	IFELSE           = 0x0313
	IFIELD           = 0x05d3
	IFPRESENT        = 0x0633
	IGET             = 0x05a3
	IN               = 0x0743
	INCONC           = 0x02c3
	INFINITYN        = 0x0263
	INFINITYP        = 0x0273
	INOUT            = 0x0753
	INT2BIT          = 0x1823
	INT2CHAR         = 0x1803
	INT2ENUM         = 0x1aa3
	INT2FLOAT        = 0x1863
	INT2HEX          = 0x1833
	INT2OCT          = 0x1843
	INT2STR          = 0x1853
	INTEGER          = 0x0e63
	INTERLEAVE       = 0x0393
	ISBOUND          = 0x1ad3
	ISCHOSEN         = 0x19f3
	ISPRESENT        = 0x19e3
	ISTR             = 0x0163
	ISVALUE          = 0x1a63
	KILL             = 0x13c3
	KILLA            = 0x13d3
	KILLED           = 0x1243
	KILLED1          = 0x1253
	KILLEDA          = 0x1263
	LE               = 0x0533
	LENGTH           = 0x0643
	LENGTHOF         = 0x19b3
	LOAD             = 0x0053
	LOG              = 0x1013
	LT               = 0x0543
	MAP              = 0x10c3
	MAPT             = 0x0fc3
	MARK             = 0x00c3
	MATCH            = 0x1023
	MOD              = 0x0463
	MODULE           = 0x08e3
	MOVE             = 0x05e3
	MPAR             = 0x0a03
	MPARD            = 0x0a13
	MTC              = 0x0213
	MUL              = 0x0443
	NAME             = 0x0183
	NATLONG          = 0x0143
	NE               = 0x0553
	NEG              = 0x0423
	NIBBLES          = 0x0123
	NONE             = 0x02a3
	NOP              = 0x0003
	NOT              = 0x0493
	NOW              = 0x1343
	NULL             = 0x0203
	OCT2BIT          = 0x1943
	OCT2CHR          = 0x1973
	OCT2HEX          = 0x1953
	OCT2INT          = 0x1933
	OCT2STR          = 0x1963
	OCTETS           = 0x0133
	OCTETSTRING      = 0x0e73
	OMIT             = 0x0253
	OR               = 0x04a3
	OUT              = 0x0763
	PASS             = 0x02b3
	PATTERN          = 0x0653
	PERMITO          = 0x0783
	PERMITP          = 0x07a3
	PERMITT          = 0x0793
	PERMUTATION      = 0x0663
	PORTM            = 0x0f43
	PORTMA           = 0x0f63
	POS              = 0x0403
	RANGE            = 0x0673
	READ             = 0x1303
	RECEIVE          = 0x1443
	RECEIVE1         = 0x1463
	RECEIVEC         = 0x1453
	RECEIVEC1        = 0x1473
	RECORD           = 0x0f83
	RECORDOF         = 0x0fa3
	REF_BIT2HEX      = 0x1dc3
	REF_BIT2INT      = 0x1db3
	REF_BIT2OCT      = 0x1dd3
	REF_BIT2STR      = 0x1de3
	REF_CHAR2INT     = 0x1d83
	REF_CHAR2OCT     = 0x1d93
	REF_DECVALUE     = 0x1f83
	REF_ENCVALUE     = 0x1f73
	REF_ENUM2INT     = 0x1f53
	REF_FLOAT2INT    = 0x1d73
	REF_HEX2BIT      = 0x1e03
	REF_HEX2INT      = 0x1df3
	REF_HEX2OCT      = 0x1e13
	REF_HEX2STR      = 0x1e23
	REF_INT2BIT      = 0x1d23
	REF_INT2CHAR     = 0x1d03
	REF_INT2ENUM     = 0x1fa3
	REF_INT2FLOAT    = 0x1d63
	REF_INT2HEX      = 0x1d33
	REF_INT2OCT      = 0x1d43
	REF_INT2STR      = 0x1d53
	REF_ISBOUND      = 0x1fd3
	REF_ISPRESENT    = 0x1ee3
	REF_ISVALUE      = 0x1f63
	REF_OCT2BIT      = 0x1e43
	REF_OCT2CHR      = 0x1e73
	REF_OCT2HEX      = 0x1e53
	REF_OCT2INT      = 0x1e33
	REF_OCT2STR      = 0x1e63
	REF_REGEXP       = 0x1f03
	REF_STR2FLOAT    = 0x1ea3
	REF_STR2HEX      = 0x1fc3
	REF_STR2INT      = 0x1e83
	REF_STR2OCT      = 0x1e93
	REF_TESTCASENAME = 0x1f93
	REF_VAL2STR      = 0x1ed3
	REGEXP           = 0x1a03
	REM              = 0x0473
	REPEAT           = 0x03b3
	REPLACE          = 0x1a23
	RETURN           = 0x0323
	RND              = 0x1a33
	ROL              = 0x04e3
	ROR              = 0x04f3
	RUNNING          = 0x1203
	RUNNING1C        = 0x1213
	RUNNING1T        = 0x1313
	RUNNINGAC        = 0x1223
	SCAN             = 0x00d3
	SELF             = 0x0223
	SEND             = 0x1403
	SEND1            = 0x1413
	SENDA            = 0x1433
	SENDER           = 0x17e3
	SENDN            = 0x1423
	SET              = 0x0f93
	SETOF            = 0x0fb3
	SETVERDICT       = 0x1053
	SHL              = 0x04c3
	SHR              = 0x04d3
	SIZEOF           = 0x19c3
	SKIP             = 0x0243
	SMATCH           = 0x00f3
	SOURCE           = 0x08d3
	SPECPLC          = 0x0d43
	START            = 0x1283
	STARTAP          = 0x12b3
	STARTC           = 0x1293
	STARTD           = 0x12a3
	STEP             = 0x03e3
	STOP             = 0x12c3
	STOPAC           = 0x12d3
	STOPAP           = 0x12e3
	STOPAT           = 0x12f3
	STOPI            = 0x0333
	STORE            = 0x0063
	STR2FLOAT        = 0x19a3
	STR2HEX          = 0x1ac3
	STR2INT          = 0x1983
	STR2OCT          = 0x1993
	SUB              = 0x0433
	SUBSET           = 0x0683
	SUBSTR           = 0x1a13
	SUBTYPE          = 0x0ef3
	SUPERSET         = 0x0693
	SYSTEM           = 0x0233
	TCSTOP           = 0x11e3
	TEMPLATE         = 0x0a43
	TERM             = 0x0713
	TESTCASE         = 0x0a83
	TESTCASENAME     = 0x1a93
	TESTCASES        = 0x0a93
	TESTCASESW       = 0x0ab3
	TESTCASEW        = 0x0aa3
	TIMEOUT          = 0x1323
	TIMEOUT1         = 0x1333
	TIMER            = 0x0ed3
	TIMESTAMP        = 0x17f3
	TO               = 0x06d3
	TRIGGER          = 0x1483
	TRIGGER1         = 0x1493
	TRUE             = 0x0293
	TYPE             = 0x0a63
	TYPEW            = 0x0a73
	UNION            = 0x0ff3
	UNMAP            = 0x10d3
	UNMAPA           = 0x10e3
	UNMAPAA          = 0x10f3
	UNMAPFROMTO      = 0x0fd3
	UTF8             = 0x0103
	VAL2STR          = 0x1ae3
	VALUE            = 0x17c3
	VALUEOF          = 0x1033
	VAR              = 0x0703
	VARDUP           = 0x06f3
	VARIANT          = 0x0cc3
	VARIANTO         = 0x0ce3
	VERDICTTYPE      = 0x0e83
	VERSION          = 0x0033
	VLIST            = 0x0073
	WAIT             = 0x1353
	WHILE            = 0x0353
	XOR              = 0x04b3
)

var opcodeStrings = map[Opcode]string{
	REF:        "ref",
	LINE:       "line",
	GOTO:       "goto",
	FROZEN_REF: "frozen_ref",

	ACTION:           "action",
	ACTIVATE:         "activate",
	ADD:              "add",
	ADDRESS:          "address",
	ALIVE:            "alive",
	ALIVE1:           "alive1",
	ALIVEA:           "alivea",
	ALLFROM:          "allfrom",
	ALLFROMP:         "allfromp",
	ALT:              "alt",
	ALTND:            "altnd",
	ALTSTEP:          "altstep",
	ALTSTEPB:         "altstepb",
	ALTSTEPBW:        "altstepbw",
	ALTSTEPW:         "altstepw",
	AND:              "and",
	ANY:              "any",
	ANYN:             "anyn",
	APPLY:            "apply",
	ARRAY:            "array",
	ASSIGN:           "assign",
	ASSIGND:          "assignd",
	AT_DEFAULT:       "at_default",
	BIT2HEX:          "bit2hex",
	BIT2INT:          "bit2int",
	BIT2OCT:          "bit2oct",
	BIT2STR:          "bit2str",
	BITS:             "bits",
	BITSTRING:        "bitstring",
	BLOCK:            "block",
	BOOLEAN:          "boolean",
	BREAK:            "break",
	CAT:              "cat",
	CHAR2INT:         "char2int",
	CHAR2OCT:         "char2oct",
	CHARSTRING:       "charstring",
	CHARSTRINGU:      "charstringu",
	CHECK:            "check",
	CHECK1:           "check1",
	CHECKSTATE:       "checkstate",
	CHECKSTATEAL:     "checkstateal",
	CHECKSTATEAN:     "checkstatean",
	CLEAR:            "clear",
	CLEARA:           "cleara",
	CLOSURE:          "closure",
	CLOSURETYPE:      "closuretype",
	COLLECT:          "collect",
	COMPLEMENT:       "complement",
	COMPONENT:        "component",
	COMPONENTX:       "componentx",
	CONNECT:          "connect",
	CONST:            "const",
	CONSTW:           "constw",
	CONTINUE:         "continue",
	CONTROL:          "control",
	CREATE:           "create",
	CREATEA:          "createa",
	CREATEAN:         "createan",
	CREATEN:          "createn",
	DEACTIVATE:       "deactivate",
	DEACTIVATEA:      "deactivatea",
	DECVALUE:         "decvalue",
	DEF:              "def",
	DEFAULT:          "default",
	DISCONNECT:       "disconnect",
	DISCONNECTA:      "disconnecta",
	DISCONNECTAA:     "disconnectaa",
	DIV:              "div",
	DONE:             "done",
	DONE1:            "done1",
	DONEA:            "donea",
	DOWHILE:          "dowhile",
	DROP:             "drop",
	ELSE:             "else",
	ENCODE:           "encode",
	ENCODEO:          "encodeo",
	ENCVALUE:         "encvalue",
	ENUM2INT:         "enum2int",
	ENUMERATED:       "enumerated",
	EQ:               "eq",
	ERROR:            "error",
	EXEC:             "exec",
	EXECUTE:          "execute",
	EXECUTED:         "executed",
	EXECUTEL:         "executel",
	EXECUTELD:        "executeld",
	EXTENSION:        "extension",
	EXTENSIONO:       "extensiono",
	FAIL:             "fail",
	FALSE:            "false",
	FIELD:            "field",
	FIELDO:           "fieldo",
	FLOAT:            "float",
	FLOAT2INT:        "float2int",
	FOR:              "for",
	FROM:             "from",
	FSTR:             "fstr",
	FUNCTION:         "function",
	FUNCTIONB:        "functionb",
	FUNCTIONV:        "functionv",
	FUNCTIONVB:       "functionvb",
	FUNCTIONX:        "functionx",
	FUNCTIONXV:       "functionxv",
	FUNCTIONXVW:      "functionxvw",
	FUNCTIONXW:       "functionxw",
	GE:               "ge",
	GET:              "get",
	GETVERDICT:       "getverdict",
	GT:               "gt",
	HALT:             "halt",
	HALTA:            "halta",
	HEX2BIT:          "hex2bit",
	HEX2INT:          "hex2int",
	HEX2OCT:          "hex2oct",
	HEX2STR:          "hex2str",
	HEXSTRING:        "hexstring",
	IDEF:             "idef",
	IEEE754DP:        "ieee754dp",
	IF:               "if",
	IFELSE:           "ifelse",
	IFIELD:           "ifield",
	IFPRESENT:        "ifpresent",
	IGET:             "iget",
	IN:               "in",
	INCONC:           "inconc",
	INFINITYN:        "infinityn",
	INFINITYP:        "infinityp",
	INOUT:            "inout",
	INT2BIT:          "int2bit",
	INT2CHAR:         "int2char",
	INT2ENUM:         "int2enum",
	INT2FLOAT:        "int2float",
	INT2HEX:          "int2hex",
	INT2OCT:          "int2oct",
	INT2STR:          "int2str",
	INTEGER:          "integer",
	INTERLEAVE:       "interleave",
	ISBOUND:          "isbound",
	ISCHOSEN:         "ischosen",
	ISPRESENT:        "ispresent",
	ISTR:             "istr",
	ISVALUE:          "isvalue",
	KILL:             "kill",
	KILLA:            "killa",
	KILLED:           "killed",
	KILLED1:          "killed1",
	KILLEDA:          "killeda",
	LE:               "le",
	LENGTH:           "length",
	LENGTHOF:         "lengthof",
	LOAD:             "load",
	LOG:              "log",
	LT:               "lt",
	MAP:              "map",
	MAPT:             "mapt",
	MARK:             "mark",
	MATCH:            "match",
	MOD:              "mod",
	MODULE:           "module",
	MOVE:             "move",
	MPAR:             "mpar",
	MPARD:            "mpard",
	MTC:              "mtc",
	MUL:              "mul",
	NAME:             "name",
	NATLONG:          "natlong",
	NE:               "ne",
	NEG:              "neg",
	NIBBLES:          "nibbles",
	NONE:             "none",
	NOP:              "nop",
	NOT:              "not",
	NOW:              "now",
	NULL:             "null",
	OCT2BIT:          "oct2bit",
	OCT2CHR:          "oct2chr",
	OCT2HEX:          "oct2hex",
	OCT2INT:          "oct2int",
	OCT2STR:          "oct2str",
	OCTETS:           "octets",
	OCTETSTRING:      "octetstring",
	OMIT:             "omit",
	OR:               "or",
	OUT:              "out",
	PASS:             "pass",
	PATTERN:          "pattern",
	PERMITO:          "permito",
	PERMITP:          "permitp",
	PERMITT:          "permitt",
	PERMUTATION:      "permutation",
	PORTM:            "portm",
	PORTMA:           "portma",
	POS:              "pos",
	RANGE:            "range",
	READ:             "read",
	RECEIVE:          "receive",
	RECEIVE1:         "receive1",
	RECEIVEC:         "receivec",
	RECEIVEC1:        "receivec1",
	RECORD:           "record",
	RECORDOF:         "recordof",
	REF_BIT2HEX:      "ref_bit2hex",
	REF_BIT2INT:      "ref_bit2int",
	REF_BIT2OCT:      "ref_bit2oct",
	REF_BIT2STR:      "ref_bit2str",
	REF_CHAR2INT:     "ref_char2int",
	REF_CHAR2OCT:     "ref_char2oct",
	REF_DECVALUE:     "ref_decvalue",
	REF_ENCVALUE:     "ref_encvalue",
	REF_ENUM2INT:     "ref_enum2int",
	REF_FLOAT2INT:    "ref_float2int",
	REF_HEX2BIT:      "ref_hex2bit",
	REF_HEX2INT:      "ref_hex2int",
	REF_HEX2OCT:      "ref_hex2oct",
	REF_HEX2STR:      "ref_hex2str",
	REF_INT2BIT:      "ref_int2bit",
	REF_INT2CHAR:     "ref_int2char",
	REF_INT2ENUM:     "ref_int2enum",
	REF_INT2FLOAT:    "ref_int2float",
	REF_INT2HEX:      "ref_int2hex",
	REF_INT2OCT:      "ref_int2oct",
	REF_INT2STR:      "ref_int2str",
	REF_ISBOUND:      "ref_isbound",
	REF_ISPRESENT:    "ref_ispresent",
	REF_ISVALUE:      "ref_isvalue",
	REF_OCT2BIT:      "ref_oct2bit",
	REF_OCT2CHR:      "ref_oct2chr",
	REF_OCT2HEX:      "ref_oct2hex",
	REF_OCT2INT:      "ref_oct2int",
	REF_OCT2STR:      "ref_oct2str",
	REF_REGEXP:       "ref_regexp",
	REF_STR2FLOAT:    "ref_str2float",
	REF_STR2HEX:      "ref_str2hex",
	REF_STR2INT:      "ref_str2int",
	REF_STR2OCT:      "ref_str2oct",
	REF_TESTCASENAME: "ref_testcasename",
	REF_VAL2STR:      "ref_val2str",
	REGEXP:           "regexp",
	REM:              "rem",
	REPEAT:           "repeat",
	REPLACE:          "replace",
	RETURN:           "return",
	RND:              "rnd",
	ROL:              "rol",
	ROR:              "ror",
	RUNNING:          "running",
	RUNNING1C:        "running1c",
	RUNNING1T:        "running1t",
	RUNNINGAC:        "runningac",
	SCAN:             "scan",
	SELF:             "self",
	SEND:             "send",
	SEND1:            "send1",
	SENDA:            "senda",
	SENDER:           "sender",
	SENDN:            "sendn",
	SET:              "set",
	SETOF:            "setof",
	SETVERDICT:       "setverdict",
	SHL:              "shl",
	SHR:              "shr",
	SIZEOF:           "sizeof",
	SKIP:             "skip",
	SMATCH:           "smatch",
	SOURCE:           "source",
	SPECPLC:          "specplc",
	START:            "start",
	STARTAP:          "startap",
	STARTC:           "startc",
	STARTD:           "startd",
	STEP:             "step",
	STOP:             "stop",
	STOPAC:           "stopac",
	STOPAP:           "stopap",
	STOPAT:           "stopat",
	STOPI:            "stopi",
	STORE:            "store",
	STR2FLOAT:        "str2float",
	STR2HEX:          "str2hex",
	STR2INT:          "str2int",
	STR2OCT:          "str2oct",
	SUB:              "sub",
	SUBSET:           "subset",
	SUBSTR:           "substr",
	SUBTYPE:          "subtype",
	SUPERSET:         "superset",
	SYSTEM:           "system",
	TCSTOP:           "tcstop",
	TEMPLATE:         "template",
	TERM:             "term",
	TESTCASE:         "testcase",
	TESTCASENAME:     "testcasename",
	TESTCASES:        "testcases",
	TESTCASESW:       "testcasesw",
	TESTCASEW:        "testcasew",
	TIMEOUT:          "timeout",
	TIMEOUT1:         "timeout1",
	TIMER:            "timer",
	TIMESTAMP:        "timestamp",
	TO:               "to",
	TRIGGER:          "trigger",
	TRIGGER1:         "trigger1",
	TRUE:             "true",
	TYPE:             "type",
	TYPEW:            "typew",
	UNION:            "union",
	UNMAP:            "unmap",
	UNMAPA:           "unmapa",
	UNMAPAA:          "unmapaa",
	UNMAPFROMTO:      "unmapfromto",
	UTF8:             "utf8",
	VAL2STR:          "val2str",
	VALUE:            "value",
	VALUEOF:          "valueof",
	VAR:              "var",
	VARDUP:           "vardup",
	VARIANT:          "variant",
	VARIANTO:         "varianto",
	VERDICTTYPE:      "verdicttype",
	VERSION:          "version",
	VLIST:            "vlist",
	WAIT:             "wait",
	WHILE:            "while",
	XOR:              "xor",
}

var opcodeNames = map[string]Opcode{
	"ref":        REF,
	"line":       LINE,
	"goto":       GOTO,
	"frozen_ref": FROZEN_REF,

	"action":           ACTION,
	"activate":         ACTIVATE,
	"add":              ADD,
	"address":          ADDRESS,
	"alive":            ALIVE,
	"alive1":           ALIVE1,
	"alivea":           ALIVEA,
	"allfrom":          ALLFROM,
	"allfromp":         ALLFROMP,
	"alt":              ALT,
	"altnd":            ALTND,
	"altstep":          ALTSTEP,
	"altstepb":         ALTSTEPB,
	"altstepbw":        ALTSTEPBW,
	"altstepw":         ALTSTEPW,
	"and":              AND,
	"any":              ANY,
	"anyn":             ANYN,
	"apply":            APPLY,
	"array":            ARRAY,
	"assign":           ASSIGN,
	"assignd":          ASSIGND,
	"at_default":       AT_DEFAULT,
	"bit2hex":          BIT2HEX,
	"bit2int":          BIT2INT,
	"bit2oct":          BIT2OCT,
	"bit2str":          BIT2STR,
	"bits":             BITS,
	"bitstring":        BITSTRING,
	"block":            BLOCK,
	"boolean":          BOOLEAN,
	"break":            BREAK,
	"cat":              CAT,
	"char2int":         CHAR2INT,
	"char2oct":         CHAR2OCT,
	"charstring":       CHARSTRING,
	"charstringu":      CHARSTRINGU,
	"check":            CHECK,
	"check1":           CHECK1,
	"checkstate":       CHECKSTATE,
	"checkstateal":     CHECKSTATEAL,
	"checkstatean":     CHECKSTATEAN,
	"clear":            CLEAR,
	"cleara":           CLEARA,
	"closure":          CLOSURE,
	"closuretype":      CLOSURETYPE,
	"collect":          COLLECT,
	"complement":       COMPLEMENT,
	"component":        COMPONENT,
	"componentx":       COMPONENTX,
	"connect":          CONNECT,
	"const":            CONST,
	"constw":           CONSTW,
	"continue":         CONTINUE,
	"control":          CONTROL,
	"create":           CREATE,
	"createa":          CREATEA,
	"createan":         CREATEAN,
	"createn":          CREATEN,
	"deactivate":       DEACTIVATE,
	"deactivatea":      DEACTIVATEA,
	"decvalue":         DECVALUE,
	"def":              DEF,
	"default":          DEFAULT,
	"disconnect":       DISCONNECT,
	"disconnecta":      DISCONNECTA,
	"disconnectaa":     DISCONNECTAA,
	"div":              DIV,
	"done":             DONE,
	"done1":            DONE1,
	"donea":            DONEA,
	"dowhile":          DOWHILE,
	"drop":             DROP,
	"else":             ELSE,
	"encode":           ENCODE,
	"encodeo":          ENCODEO,
	"encvalue":         ENCVALUE,
	"enum2int":         ENUM2INT,
	"enumerated":       ENUMERATED,
	"eq":               EQ,
	"error":            ERROR,
	"exec":             EXEC,
	"execute":          EXECUTE,
	"executed":         EXECUTED,
	"executel":         EXECUTEL,
	"executeld":        EXECUTELD,
	"extension":        EXTENSION,
	"extensiono":       EXTENSIONO,
	"fail":             FAIL,
	"false":            FALSE,
	"field":            FIELD,
	"fieldo":           FIELDO,
	"float":            FLOAT,
	"float2int":        FLOAT2INT,
	"for":              FOR,
	"from":             FROM,
	"fstr":             FSTR,
	"function":         FUNCTION,
	"functionb":        FUNCTIONB,
	"functionv":        FUNCTIONV,
	"functionvb":       FUNCTIONVB,
	"functionx":        FUNCTIONX,
	"functionxv":       FUNCTIONXV,
	"functionxvw":      FUNCTIONXVW,
	"functionxw":       FUNCTIONXW,
	"ge":               GE,
	"get":              GET,
	"getverdict":       GETVERDICT,
	"gt":               GT,
	"halt":             HALT,
	"halta":            HALTA,
	"hex2bit":          HEX2BIT,
	"hex2int":          HEX2INT,
	"hex2oct":          HEX2OCT,
	"hex2str":          HEX2STR,
	"hexstring":        HEXSTRING,
	"idef":             IDEF,
	"ieee754dp":        IEEE754DP,
	"if":               IF,
	"ifelse":           IFELSE,
	"ifield":           IFIELD,
	"ifpresent":        IFPRESENT,
	"iget":             IGET,
	"in":               IN,
	"inconc":           INCONC,
	"infinityn":        INFINITYN,
	"infinityp":        INFINITYP,
	"inout":            INOUT,
	"int2bit":          INT2BIT,
	"int2char":         INT2CHAR,
	"int2enum":         INT2ENUM,
	"int2float":        INT2FLOAT,
	"int2hex":          INT2HEX,
	"int2oct":          INT2OCT,
	"int2str":          INT2STR,
	"integer":          INTEGER,
	"interleave":       INTERLEAVE,
	"isbound":          ISBOUND,
	"ischosen":         ISCHOSEN,
	"ispresent":        ISPRESENT,
	"istr":             ISTR,
	"isvalue":          ISVALUE,
	"kill":             KILL,
	"killa":            KILLA,
	"killed":           KILLED,
	"killed1":          KILLED1,
	"killeda":          KILLEDA,
	"le":               LE,
	"length":           LENGTH,
	"lengthof":         LENGTHOF,
	"load":             LOAD,
	"log":              LOG,
	"lt":               LT,
	"map":              MAP,
	"mapt":             MAPT,
	"mark":             MARK,
	"match":            MATCH,
	"mod":              MOD,
	"module":           MODULE,
	"move":             MOVE,
	"mpar":             MPAR,
	"mpard":            MPARD,
	"mtc":              MTC,
	"mul":              MUL,
	"name":             NAME,
	"natlong":          NATLONG,
	"ne":               NE,
	"neg":              NEG,
	"nibbles":          NIBBLES,
	"none":             NONE,
	"nop":              NOP,
	"not":              NOT,
	"now":              NOW,
	"null":             NULL,
	"oct2bit":          OCT2BIT,
	"oct2chr":          OCT2CHR,
	"oct2hex":          OCT2HEX,
	"oct2int":          OCT2INT,
	"oct2str":          OCT2STR,
	"octets":           OCTETS,
	"octetstring":      OCTETSTRING,
	"omit":             OMIT,
	"or":               OR,
	"out":              OUT,
	"pass":             PASS,
	"pattern":          PATTERN,
	"permito":          PERMITO,
	"permitp":          PERMITP,
	"permitt":          PERMITT,
	"permutation":      PERMUTATION,
	"portm":            PORTM,
	"portma":           PORTMA,
	"pos":              POS,
	"range":            RANGE,
	"read":             READ,
	"receive":          RECEIVE,
	"receive1":         RECEIVE1,
	"receivec":         RECEIVEC,
	"receivec1":        RECEIVEC1,
	"record":           RECORD,
	"recordof":         RECORDOF,
	"ref_bit2hex":      REF_BIT2HEX,
	"ref_bit2int":      REF_BIT2INT,
	"ref_bit2oct":      REF_BIT2OCT,
	"ref_bit2str":      REF_BIT2STR,
	"ref_char2int":     REF_CHAR2INT,
	"ref_char2oct":     REF_CHAR2OCT,
	"ref_decvalue":     REF_DECVALUE,
	"ref_encvalue":     REF_ENCVALUE,
	"ref_enum2int":     REF_ENUM2INT,
	"ref_float2int":    REF_FLOAT2INT,
	"ref_hex2bit":      REF_HEX2BIT,
	"ref_hex2int":      REF_HEX2INT,
	"ref_hex2oct":      REF_HEX2OCT,
	"ref_hex2str":      REF_HEX2STR,
	"ref_int2bit":      REF_INT2BIT,
	"ref_int2char":     REF_INT2CHAR,
	"ref_int2enum":     REF_INT2ENUM,
	"ref_int2float":    REF_INT2FLOAT,
	"ref_int2hex":      REF_INT2HEX,
	"ref_int2oct":      REF_INT2OCT,
	"ref_int2str":      REF_INT2STR,
	"ref_isbound":      REF_ISBOUND,
	"ref_ispresent":    REF_ISPRESENT,
	"ref_isvalue":      REF_ISVALUE,
	"ref_oct2bit":      REF_OCT2BIT,
	"ref_oct2chr":      REF_OCT2CHR,
	"ref_oct2hex":      REF_OCT2HEX,
	"ref_oct2int":      REF_OCT2INT,
	"ref_oct2str":      REF_OCT2STR,
	"ref_regexp":       REF_REGEXP,
	"ref_str2float":    REF_STR2FLOAT,
	"ref_str2hex":      REF_STR2HEX,
	"ref_str2int":      REF_STR2INT,
	"ref_str2oct":      REF_STR2OCT,
	"ref_testcasename": REF_TESTCASENAME,
	"ref_val2str":      REF_VAL2STR,
	"regexp":           REGEXP,
	"rem":              REM,
	"repeat":           REPEAT,
	"replace":          REPLACE,
	"return":           RETURN,
	"rnd":              RND,
	"rol":              ROL,
	"ror":              ROR,
	"running":          RUNNING,
	"running1c":        RUNNING1C,
	"running1t":        RUNNING1T,
	"runningac":        RUNNINGAC,
	"scan":             SCAN,
	"self":             SELF,
	"send":             SEND,
	"send1":            SEND1,
	"senda":            SENDA,
	"sender":           SENDER,
	"sendn":            SENDN,
	"set":              SET,
	"setof":            SETOF,
	"setverdict":       SETVERDICT,
	"shl":              SHL,
	"shr":              SHR,
	"sizeof":           SIZEOF,
	"skip":             SKIP,
	"smatch":           SMATCH,
	"source":           SOURCE,
	"specplc":          SPECPLC,
	"start":            START,
	"startap":          STARTAP,
	"startc":           STARTC,
	"startd":           STARTD,
	"step":             STEP,
	"stop":             STOP,
	"stopac":           STOPAC,
	"stopap":           STOPAP,
	"stopat":           STOPAT,
	"stopi":            STOPI,
	"store":            STORE,
	"str2float":        STR2FLOAT,
	"str2hex":          STR2HEX,
	"str2int":          STR2INT,
	"str2oct":          STR2OCT,
	"sub":              SUB,
	"subset":           SUBSET,
	"substr":           SUBSTR,
	"subtype":          SUBTYPE,
	"superset":         SUPERSET,
	"system":           SYSTEM,
	"tcstop":           TCSTOP,
	"template":         TEMPLATE,
	"term":             TERM,
	"testcase":         TESTCASE,
	"testcasename":     TESTCASENAME,
	"testcases":        TESTCASES,
	"testcasesw":       TESTCASESW,
	"testcasew":        TESTCASEW,
	"timeout":          TIMEOUT,
	"timeout1":         TIMEOUT1,
	"timer":            TIMER,
	"timestamp":        TIMESTAMP,
	"to":               TO,
	"trigger":          TRIGGER,
	"trigger1":         TRIGGER1,
	"true":             TRUE,
	"type":             TYPE,
	"typew":            TYPEW,
	"union":            UNION,
	"unmap":            UNMAP,
	"unmapa":           UNMAPA,
	"unmapaa":          UNMAPAA,
	"unmapfromto":      UNMAPFROMTO,
	"utf8":             UTF8,
	"val2str":          VAL2STR,
	"value":            VALUE,
	"valueof":          VALUEOF,
	"var":              VAR,
	"vardup":           VARDUP,
	"variant":          VARIANT,
	"varianto":         VARIANTO,
	"verdicttype":      VERDICTTYPE,
	"version":          VERSION,
	"vlist":            VLIST,
	"wait":             WAIT,
	"while":            WHILE,
	"xor":              XOR,
}
