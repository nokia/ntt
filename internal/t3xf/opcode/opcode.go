package opcode

import "fmt"

const (
	refClass   = 0
	gotoClass  = 1
	lineClass  = 2
	instrClass = 3
)

const (
	NOP              Opcode = 0x000
	ESWAP                   = 0x001
	WIDEN                   = 0x002
	VERSION                 = 0x003
	VLIST                   = 0x007
	BLOCK                   = 0x008
	COLLECT                 = 0x009
	DROP                    = 0x00a
	MARK                    = 0x00c
	SCAN                    = 0x00d
	SMATCH                  = 0x00f
	UTF8                    = 0x010
	BITS                    = 0x011
	NIBBLES                 = 0x012
	OCTETS                  = 0x013
	NATLONG                 = 0x014
	IEEE754DP               = 0x015
	ISTR                    = 0x016
	FSTR                    = 0x017
	NAME                    = 0x018
	APPLY                   = 0x019
	OPNULL                  = 0x020
	MTC                     = 0x021
	SELF                    = 0x022
	SYSTEM                  = 0x023
	SKIP                    = 0x024
	OMIT                    = 0x025
	INFINITYN               = 0x026
	INFINITYP               = 0x027
	FALSE                   = 0x028
	TRUE                    = 0x029
	NONE                    = 0x02a
	PASS                    = 0x02b
	INCONC                  = 0x02c
	FAIL                    = 0x02d
	ERROR                   = 0x02e
	IF                      = 0x030
	IFELSE                  = 0x031
	RETURN                  = 0x032
	STOPI                   = 0x033
	FOR                     = 0x034
	WHILE                   = 0x035
	DOWHILE                 = 0x036
	EXEC                    = 0x037
	ALT                     = 0x038
	INTERLEAVE              = 0x039
	ELSE                    = 0x03a
	REPEAT                  = 0x03b
	BREAK                   = 0x03c
	CONTINUE                = 0x03d
	STEP                    = 0x03e
	ACACHE                  = 0x03f
	POS                     = 0x040
	ADD                     = 0x041
	NEG                     = 0x042
	SUB                     = 0x043
	MUL                     = 0x044
	DIV                     = 0x045
	MOD                     = 0x046
	REM                     = 0x047
	AND                     = 0x048
	NOT                     = 0x049
	OR                      = 0x04a
	XOR                     = 0x04b
	SHL                     = 0x04c
	SHR                     = 0x04d
	ROL                     = 0x04e
	ROR                     = 0x04f
	EQ                      = 0x050
	GE                      = 0x051
	GT                      = 0x052
	LE                      = 0x053
	LT                      = 0x054
	NE                      = 0x055
	CAT                     = 0x056
	GET                     = 0x057
	ASSIGN                  = 0x058
	ASSIGND                 = 0x059
	IGET                    = 0x05a
	DEF                     = 0x05b
	IDEF                    = 0x05c
	IFIELD                  = 0x05d
	MOVE                    = 0x05e
	ANY                     = 0x060
	ANYN                    = 0x061
	COMPLEMENT              = 0x062
	IFPRESENT               = 0x063
	LENGTH                  = 0x064
	PATTERN                 = 0x065
	PERMUTATION             = 0x066
	RANGE                   = 0x067
	SUBSET                  = 0x068
	SUPERSET                = 0x069
	ALLFROM                 = 0x06a
	ALLFROMP                = 0x06b
	VARDUP                  = 0x06f
	VAR                     = 0x070
	TERM                    = 0x071
	FIELD                   = 0x072
	FIELDO                  = 0x073
	IN                      = 0x074
	INOUT                   = 0x075
	OUT                     = 0x076
	PERMITO                 = 0x078
	PERMITT                 = 0x079
	PERMITP                 = 0x07a
	ALTSTEP                 = 0x080
	ALTSTEPB                = 0x081
	ALTSTEPW                = 0x082
	ALTSTEPBW               = 0x083
	CONST                   = 0x084
	CONSTW                  = 0x085
	CONTROL                 = 0x086
	CONTROLW                = 0x087
	GROUP                   = 0x088
	GROUPW                  = 0x089
	IMPORT                  = 0x08a
	IMPORTW                 = 0x08b
	EXTERN                  = 0x08c
	SOURCE                  = 0x08d
	MODULE                  = 0x08e
	MODULEW                 = 0x08f
	FUNCTION                = 0x090
	FUNCTIONB               = 0x091
	FUNCTIONW               = 0x092
	FUNCTIONBW              = 0x093
	FUNCTIONV               = 0x094
	FUNCTIONVB              = 0x095
	FUNCTIONVW              = 0x096
	FUNCTIONVBW             = 0x097
	FUNCTIONX               = 0x098
	FUNCTIONXW              = 0x099
	FUNCTIONXV              = 0x09a
	FUNCTIONXVW             = 0x09b
	MPAR                    = 0x0a0
	MPARD                   = 0x0a1
	MPARW                   = 0x0a2
	MPARDW                  = 0x0a3
	TEMPLATE                = 0x0a4
	TEMPLATEW               = 0x0a5
	TYPE                    = 0x0a6
	TYPEW                   = 0x0a7
	TESTCASE                = 0x0a8
	TESTCASES               = 0x0a9
	TESTCASEW               = 0x0aa
	TESTCASESW              = 0x0ab
	SIG                     = 0x0b0
	SIGA                    = 0x0b1
	SIGW                    = 0x0b2
	SIGAW                   = 0x0b3
	SIGV                    = 0x0b4
	SIGVW                   = 0x0b5
	SIGX                    = 0x0b8
	SIGXA                   = 0x0b9
	SIGXW                   = 0x0ba
	SIGXAW                  = 0x0bb
	SIGXV                   = 0x0bc
	SIGXVW                  = 0x0bd
	DISPLAY                 = 0x0c0
	DISPLAYQ                = 0x0c1
	DISPLAYO                = 0x0c2
	DISPLAYQO               = 0x0c3
	ENCODE                  = 0x0c4
	ENCODEQ                 = 0x0c5
	ENCODEO                 = 0x0c6
	ENCODEQO                = 0x0c7
	EXTENSION               = 0x0c8
	EXTENSIONQ              = 0x0c9
	EXTENSIONO              = 0x0ca
	EXTENSIONQO             = 0x0cb
	VARIANT                 = 0x0cc
	VARIANTQ                = 0x0cd
	VARIANTO                = 0x0ce
	VARIANTQO               = 0x0cf
	ACTIVATE                = 0x0d0
	DEACTIVATE              = 0x0d2
	DEACTIVATEA             = 0x0d3
	SPECPLC                 = 0x0d4
	BITSTRING               = 0x0e0
	BOOLEAN                 = 0x0e1
	CHARSTRING              = 0x0e2
	CHARSTRINGU             = 0x0e3
	FLOAT                   = 0x0e4
	HEXSTRING               = 0x0e5
	INTEGER                 = 0x0e6
	OCTETSTRING             = 0x0e7
	VERDICTTYPE             = 0x0e8
	CLOSURETYPE             = 0x0e9
	ADDRESS                 = 0x0ea
	DEFAULT                 = 0x0ec
	TIMER                   = 0x0ed
	ARRAY                   = 0x0ee
	SUBTYPE                 = 0x0ef
	ENUMERATED              = 0x0f0
	COMPONENT               = 0x0f2
	COMPONENTX              = 0x0f3
	PORTM                   = 0x0f4
	PORTP                   = 0x0f5
	PORTMA                  = 0x0f6
	PORTPA                  = 0x0f7
	RECORD                  = 0x0f8
	SET                     = 0x0f9
	RECORDOF                = 0x0fa
	SETOF                   = 0x0fb
	UNION                   = 0x0ff
	ACTION                  = 0x100
	LOG                     = 0x101
	MATCH                   = 0x102
	VALUEOF                 = 0x103
	GETVERDICT              = 0x104
	SETVERDICT              = 0x105
	EXECUTE                 = 0x106
	EXECUTEL                = 0x107
	CONNECT                 = 0x108
	DISCONNECT              = 0x109
	DISCONNECTA             = 0x10a
	DISCONNECTAA            = 0x10b
	MAP                     = 0x10c
	UNMAP                   = 0x10d
	UNMAPA                  = 0x10e
	UNMAPAA                 = 0x10f
	ALIVE                   = 0x110
	ALIVE1                  = 0x111
	ALIVEA                  = 0x112
	DONE                    = 0x114
	DONE1                   = 0x115
	DONEA                   = 0x116
	CREATE                  = 0x118
	CREATEA                 = 0x119
	CREATEN                 = 0x11a
	CREATEAN                = 0x11b
	EXECUTELD               = 0x11c
	EXECUTED                = 0x11d
	TCSTOP                  = 0x11e
	CLOSURE                 = 0x11f
	RUNNING                 = 0x120
	RUNNING1C               = 0x121
	RUNNINGAC               = 0x122
	KILLED                  = 0x124
	KILLED1                 = 0x125
	KILLEDA                 = 0x126
	START                   = 0x128
	STARTC                  = 0x129
	STARTD                  = 0x12a
	STARTAP                 = 0x12b
	STOP                    = 0x12c
	STOPAC                  = 0x12d
	STOPAP                  = 0x12e
	STOPAT                  = 0x12f
	READ                    = 0x130
	RUNNING1T               = 0x131
	TIMEOUT                 = 0x132
	TIMEOUT1                = 0x133
	NOW                     = 0x134
	WAIT                    = 0x135
	CLEAR                   = 0x138
	CLEARA                  = 0x139
	HALT                    = 0x13a
	HALTA                   = 0x13b
	KILL                    = 0x13c
	KILLA                   = 0x13d
	SEND                    = 0x140
	SEND1                   = 0x141
	SENDN                   = 0x142
	SENDA                   = 0x143
	RECEIVE                 = 0x144
	RECEIVEC                = 0x145
	RECEIVE1                = 0x146
	RECEIVEC1               = 0x147
	TRIGGER                 = 0x148
	TRIGGER1                = 0x149
	CHECK                   = 0x14a
	CHECK1                  = 0x14b
	MAP3                    = 0x14c
	MAP4                    = 0x14d
	CALL                    = 0x150
	CALL1                   = 0x151
	CALLN                   = 0x152
	CALLA                   = 0x153
	CALLB                   = 0x154
	CALLB1                  = 0x155
	CALLBN                  = 0x156
	CALLBA                  = 0x157
	CALLW                   = 0x158
	CALLW1                  = 0x159
	CALLWN                  = 0x15a
	CALLWA                  = 0x15b
	GETCALL                 = 0x15c
	GETCALLC                = 0x15d
	GETCALL1                = 0x15e
	GETCALLC1               = 0x15f
	REPLY                   = 0x160
	REPLY1                  = 0x161
	REPLYN                  = 0x162
	REPLYA                  = 0x163
	REPLYV                  = 0x164
	REPLYV1                 = 0x165
	REPLYVN                 = 0x166
	REPLYVA                 = 0x167
	GETREPLY                = 0x168
	GETREPLYC               = 0x169
	GETREPLY1               = 0x16a
	GETREPLYC1              = 0x16b
	RAISE                   = 0x170
	RAISE1                  = 0x171
	RAISEN                  = 0x172
	RAISEA                  = 0x173
	CATCH                   = 0x174
	CATCHC                  = 0x175
	CATCH1                  = 0x176
	CATCHC1                 = 0x177
	CHECKSTATE              = 0x178
	CHECKSTATEAL            = 0x179
	CHECKSTATEAN            = 0x17a
	VALUE                   = 0x17c
	PARAM                   = 0x17d
	SENDER                  = 0x17e
	TIMESTAMP               = 0x17f
	INT2CHAR                = 0x180
	INT2UNICHAR             = 0x181
	INT2BIT                 = 0x182
	INT2HEX                 = 0x183
	INT2OCT                 = 0x184
	INT2STR                 = 0x185
	INT2FLOAT               = 0x186
	FLOAT2INT               = 0x187
	CHAR2INT                = 0x188
	CHAR2OCT                = 0x189
	UNICHAR2INT             = 0x18a
	BIT2INT                 = 0x18b
	BIT2HEX                 = 0x18c
	BIT2OCT                 = 0x18d
	BIT2STR                 = 0x18e
	HEX2INT                 = 0x18f
	HEX2BIT                 = 0x190
	HEX2OCT                 = 0x191
	HEX2STR                 = 0x192
	OCT2INT                 = 0x193
	OCT2BIT                 = 0x194
	OCT2HEX                 = 0x195
	OCT2STR                 = 0x196
	OCT2CHR                 = 0x197
	STR2INT                 = 0x198
	STR2OCT                 = 0x199
	STR2FLOAT               = 0x19a
	LENGTHOF                = 0x19b
	SIZEOF                  = 0x19c
	ISPRESENT               = 0x19e
	ISCHOSEN                = 0x19f
	REGEXP                  = 0x1a0
	SUBSTR                  = 0x1a1
	REPLACE                 = 0x1a2
	RND                     = 0x1a3
	RNDS                    = 0x1a4
	ENUM2INT                = 0x1a5
	ISVALUE                 = 0x1a6
	ENCVALUE                = 0x1a7
	DECVALUE                = 0x1a8
	TESTCASENAME            = 0x1a9
	INT2ENUM                = 0x1aa
	XINT2ENUM               = 0x1ab
	STR2HEX                 = 0x1ac
	ISBOUND                 = 0x1ad
	VAL2STR                 = 0x1ae
	REF_INT2CHAR            = 0x1d0
	REF_INT2UNICHAR         = 0x1d1
	REF_INT2BIT             = 0x1d2
	REF_INT2HEX             = 0x1d3
	REF_INT2OCT             = 0x1d4
	REF_INT2STR             = 0x1d5
	REF_INT2FLOAT           = 0x1d6
	REF_FLOAT2INT           = 0x1d7
	REF_CHAR2INT            = 0x1d8
	REF_CHAR2OCT            = 0x1d9
	REF_UNICHAR2INT         = 0x1da
	REF_BIT2INT             = 0x1db
	REF_BIT2HEX             = 0x1dc
	REF_BIT2OCT             = 0x1dd
	REF_BIT2STR             = 0x1de
	REF_HEX2INT             = 0x1df
	REF_HEX2BIT             = 0x1e0
	REF_HEX2OCT             = 0x1e1
	REF_HEX2STR             = 0x1e2
	REF_OCT2INT             = 0x1e3
	REF_OCT2BIT             = 0x1e4
	REF_OCT2HEX             = 0x1e5
	REF_OCT2STR             = 0x1e6
	REF_OCT2CHR             = 0x1e7
	REF_STR2INT             = 0x1e8
	REF_STR2OCT             = 0x1e9
	REF_STR2FLOAT           = 0x1ea
	REF_LENGTHOF            = 0x1eb
	REF_SIZEOF              = 0x1ec
	REF_VAL2STR             = 0x1ed
	REF_ISPRESENT           = 0x1ee
	REF_ISCHOSEN            = 0x1ef
	REF_REGEXP              = 0x1f0
	REF_SUBSTR              = 0x1f1
	REF_REPLACE             = 0x1f2
	REF_RND                 = 0x1f3
	REF_RNDS                = 0x1f4
	REF_ENUM2INT            = 0x1f5
	REF_ISVALUE             = 0x1f6
	REF_ENCVALUE            = 0x1f7
	REF_DECVALUE            = 0x1f8
	REF_TESTCASENAME        = 0x1f9
	REF_INT2ENUM            = 0x1fa
	REF_XINT2ENUM           = 0x1fb
	REF_STR2HEX             = 0x1fc
	REF_ISBOUND             = 0x1fd
	PROFILE_TIME            = 0x1fe
	AT_DEFAULT              = 0x201

	// Pseudo opcodes
	LINE       = -10
	GOTO       = -20
	REF        = -30
	FROZEN_REF = -40
)

type Opcode int

func (op Opcode) String() string {
	switch op {
	case LINE:
		return "line"
	case GOTO:
		return "goto"
	case REF:
		return "ref"
	case FROZEN_REF:
		return "frozen_ref"
	}

	s := ""
	if 0 <= op && op < Opcode(len(opcodeStrings)) {
		s = opcodeStrings[op]
	}
	if s == "" {
		s = fmt.Sprintf("unknown_opcode(0x%x)", int(op))
	}
	return s
}

func (op Opcode) HasArg() bool {
	switch op {
	case IDEF:
		return true
	case IGET:
		return true
	case IFIELD:
		return true
	case REF:
		return true
	case GOTO:
		return true
	case LINE:
		return true
	default:
		return false
	}
}

func Unpack(x uint32) (Opcode, int) {
	i := int(x)
	switch i & 0x3 {
	case instrClass:
		return Opcode(i >> 4 & 0xfff), i >> 16
	case refClass:
		if uint32(i)&(1<<31) != 0 {
			return FROZEN_REF, int(uint32(i) &^ (1 << 31))
		}
		return REF, i
	case lineClass:
		return LINE, i >> 2
	case gotoClass:
		return GOTO, i & ^(0x3)
	}

	panic("internal error")
}
