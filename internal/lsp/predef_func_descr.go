package lsp

import "github.com/nokia/ntt/internal/lsp/protocol"

type PredefFunctionDetails struct {
	Label          string
	InsertText     string
	Signature      string
	Documentation  string
	NrOfParameters int
	TextFormat     protocol.InsertTextFormat
}

var predefinedFunctions = []PredefFunctionDetails{
	{
		Label:          "int2char(...)",
		InsertText:     "int2char(${1:invalue})$0",
		Signature:      "int2char(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __int2char__ function converts an __integer__ value in the range of 0 to 127 (8-bit encoding) into a single-character-length\n __charstring__ value. The __integer__ value describes the 8-bit encoding of the character",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2unichar(...)",
		InsertText:     "int2unichar(${1:invalue})$0",
		Signature:      "int2unichar(in integer invalue) return universal charstring",
		Documentation:  "## (TTCN-3)\nThe __int2unichar__ function n converts an __integer__ value in the range of 0 to 2147483647 (32-bit encoding) into a single-character-length __universal charstring__ value. The __integer__ value describes the 32-bit encoding of the character.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2bit(...)",
		InsertText:     "int2bit(${1:invalue}, ${2:length})$0",
		Signature:      "int2bit(in integer invalue, in integer length) return bitstring",
		Documentation:  "## (TTCN-3)\nThe __int2bit__ function converts a single __integer__ value to a single __bitstring__ value. The resulting string is length bits long. Error causes are:\n* invalue is less than zero;\n* the conversion yields a return value with more bits than specified by length.",
		NrOfParameters: 2,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2enum(...)",
		InsertText:     "int2enum(${1:invalue}, ${2:outpar})$0",
		Signature:      "int2enum ( in integer inpar, out Enumerated_type outpar)",
		Documentation:  "## (TTCN-3)\nThe __int2enum__ function converts an integer value into an enumerated value of a given enumerated type. The integer value shall be provided as in parameter and the result of the conversion shall be stored in an out parameter. The type of the out parameter determines the type into which the in parameter is converted.",
		NrOfParameters: 2,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2hex(...)",
		InsertText:     "int2hex(${1:invalue},${2:length})$0",
		Signature:      "int2hex(in integer invalue, in integer length) return hexstring",
		Documentation:  "## (TTCN-3)\nThe __int2hex__ function converts a single __integer__ value to a single __hexstring__ value. The resulting string is length hexadecimal digits long. Error causes are:\n* invalue is less than zero;\n* the conversion yields a return value with more hexadecimal characters than specified by length.",
		NrOfParameters: 2,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2oct(...)",
		InsertText:     "int2oct(${1:invalue},${2:length})$0",
		Signature:      "int2oct(in integer invalue, in integer length) return octetstring",
		Documentation:  "## (TTCN-3)\nThe __int2oct__ function converts a single __integer__ value to a single __octetstring__ value. The resulting string is length octets long. Error causes are:\n* invalue is less than zero;\n* the conversion yields a return value with more octets than specified by length.",
		NrOfParameters: 2,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2str(...)",
		InsertText:     "int2str(${1:invalue})$0",
		Signature:      "int2str(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __int2str__ function converts the __integer__ value into its string equivalent (the base of the return string is always decimal). ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "int2float(...)",
		InsertText:     "int2float(${1:invalue})$0",
		Signature:      "int2float(in integer invalue) return float",
		Documentation:  "## (TTCN-3)\nThe __int2float__ function converts an __integer__ value into a __float__ value.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "float2int(...)",
		InsertText:     "float2int(${1:invalue})$0",
		Signature:      "float2int(in float invalue) return integer",
		Documentation:  "## (TTCN-3)\nThe __float2int__ function converts a __float__ value into an __integer__ value by removing the fractional part of the argument and returning the resulting __integer__. Error causes are:\n*invalue is __infinity__, __-infinity__ or not_a_number.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "char2int(...)",
		InsertText:     "char2int(${1:invalue})$0",
		Signature:      "char2int(in charstring invalue) return integer",
		Documentation:  "## (TTCN-3)\nThe __char2int__ function converts a single-character-length __charstring__ value into an __integer__ value in the range of 0 to 127. The __integer__ value describes the 8-bit encoding of the character. Error causes are:\n* length of invalue does not equal 1.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "char2oct(...)",
		InsertText:     "char2oct(${1:invalue})$0",
		Signature:      "char2oct(in charstring invalue) return octetstring",
		Documentation:  "## (TTCN-3)\nThe __char2oct__ function converts a __charstring__ invalue to an __octetstring__. Each octet of the octetstring will contain the Recommendation _ITU-T T.50 [4]_ codes (according to the IRV) of the appropriate characters of invalue.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "unichar2int(...)",
		InsertText: "unichar2int(${1:invalue})$0",
		Signature:  "unichar2int(in universal charstring invalue) return integer",
		Documentation: `## (TTCN-3)
The __unichar2int__ function converts a single-character-length __universal charstring__ value into an __integer__ value in the
range of 0 to 2147483647. The __integer__ value describes the 32-bit encoding of the character. Error causes are:
* length of invalue does not equal 1.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "unichar2oct(...)",
		InsertText: "unichar2oct(${1:invalue})$0",
		Signature:  "unichar2oct(in universal charstring invalue, in charstring string_encoding := \"UTF-8\") return octetstring",
		Documentation: `## (TTCN-3)
The __unichar2oct__ function This function converts a universal charstring invalue to an octetstring. Each octet of the octetstring
will contain the octets mandated by mapping the characters of invalue using the standardized mapping associated with
the given string_encoding in the same order as the characters appear in inpar. If the optional string_encoding parameter
is omitted, the default value "UTF-8".
The following values (see _ISO/IEC 10646 [2]_) are allowed as string_encoding actual parameter:

a) __"UTF-8"__ b) __"UTF-16"__ c) __"UTF-16LE"__ d) __"UTF-16BE"__ e) __"UTF-32"__ f) __"UTF-32LE"__ g) __"UTF-32BE"__`,
		NrOfParameters: 2,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "bit2int(...)",
		InsertText: "bit2int(${1:invalue})$0",
		Signature:  "bit2int(in bitstring invalue) return integer",
		Documentation: `## (TTCN-3)
The __bit2int__ function This function converts a single __bitstring__ value to a single __integer__ value.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "bit2hex(...)",
		InsertText: "bit2hex(${1:invalue})$0",
		Signature:  "bit2hex(in bitstring invalue) return hexstring",
		Documentation: `## (TTCN-3)
The __bit2hex__ function converts a single __bitstring__ value to a single __hexstring__. The resulting __hexstring__
represents the same value as the __bitstring__. For the purpose of this conversion, a __bitstring__ shall be converted into a
__hexstring__, where the __bitstring__ is divided into groups of four bits beginning with the rightmost bit.
When the leftmost group of bits does contain less than 4 bits, this group is filled with _'0'B_ from the left until it contains
exactly 4 bits and is converted afterwards. The consecutive order of hex digits in the resulting __hexstring__ is the same as
the order of groups of 4 bits in the __bitstring__`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "bit2oct(...)",
		InsertText: "bit2oct(${1:invalue})$0",
		Signature:  "bit2oct(in bitstring invalue) return octetstring",
		Documentation: `## (TTCN-3)
The __bit2oct__ function converts a single __bitstring__ value to a single __octetstring__. The resulting __octetstring__
represents the same value as the __bitstring__. For the conversion the following holds: * bit2oct(value)=hex2oct(bit2hex(value)).`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "bit2str(...)",
		InsertText: "bit2str(${1:invalue})$0",
		Signature:  "bit2str(in bitstring invalue) return charstring",
		Documentation: `## (TTCN-3)
The __bit2str__ function converts a single __bitstring__ value to a single __charstring__.
The resulting __charstring__ has the same length as the __bitstring__ and contains only the __characters__ '0' and '1'.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "hex2int(...)",
		InsertText: "hex2int(${1:invalue})$0",
		Signature:  "hex2int(in hexstring invalue) return integer",
		Documentation: `## (TTCN-3)
The __hex2int__ function  converts a single __hexstring__ value to a single __integer__ value.
For the purposes of this conversion, a __hexstring__ shall be interpreted as a positive base 16 __integer__ value. The
rightmost hexadecimal digit is least significant, the leftmost hexadecimal digit is the most significant.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "hex2bit(...)",
		InsertText: "hex2bit(${1:invalue})$0",
		Signature:  "hex2bit(in hexstring invalue) return bitstring",
		Documentation: `## (TTCN-3)
The __hex2bit__ function converts a single __hexstring__ value to a single __bitstring__. The resulting __bitstring__
 represents the same value as the __hexstring__.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "hex2oct(...)",
		InsertText: "hex2oct(${1:invalue})$0",
		Signature:  "hex2oct(in hexstring invalue) return octetstring",
		Documentation: `## (TTCN-3)
The __hex2oct__ function This function converts a single __hexstring__ value to a single __octetstring__.
The resulting __octetstring__ represents the same value as the __hexstring__.
For the purpose of this conversion, a __hexstring__ shall be converted into a __octetstring__, where the
__octetstring__ contains the same sequence of hex digits as the __hexstring__ when the length of the __hexstring__
modulo 2 is 0. Otherwise, the resulting __octetstring__ contains 0 as leftmost hex digit followed by the same sequence
of hex digits as in the __hexstring__. `,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "hex2str(...)",
		InsertText: "hex2str(${1:invalue})$0",
		Signature:  "hex2str(in hexstring invalue) return charstring",
		Documentation: `## (TTCN-3)
The __hex2str__ function converts a single __hexstring__ value to a single __charstring__. The resulting __charstring__
has the same length as the __hexstring__ and contains only the characters '0' to '9'and 'A' to 'F'.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2int(...)",
		InsertText:     "oct2int(${1:invalue})$0",
		Signature:      "oct2int(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2int__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2bit(...)",
		InsertText:     "oct2bit(${1:invalue})$0",
		Signature:      "oct2bit(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2bit__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2hex(...)",
		InsertText:     "oct2hex(${1:invalue})$0",
		Signature:      "oct2hex(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2hex__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2str(...)",
		InsertText:     "oct2str(${1:invalue})$0",
		Signature:      "oct2str(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2str__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2char(...)",
		InsertText:     "oct2char(${1:invalue})$0",
		Signature:      "oct2char(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2char__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "oct2unichar(...)",
		InsertText:     "oct2unichar(${1:invalue})$0",
		Signature:      "oct2unichar(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __oct2unichar__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "str2int(...)",
		InsertText:     "str2int(${1:invalue})$0",
		Signature:      "str2int(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __str2int__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "str2hex(...)",
		InsertText:     "str2hex(${1:invalue})$0",
		Signature:      "str2hex(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __str2hex__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "str2oct(...)",
		InsertText:     "str2oct(${1:invalue})$0",
		Signature:      "str2oct(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __str2oct__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "str2float(...)",
		InsertText:     "str2float(${1:invalue})$0",
		Signature:      "str2float(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __str2float__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "enum2int(...)",
		InsertText:     "enum2int(${1:invalue})$0",
		Signature:      "enum2int(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __enum2int__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "any2unistr(...)",
		InsertText:     "any2unistr(${1:invalue})$0",
		Signature:      "any2unistr(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __any2unistr__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "lengthof(...)",
		InsertText:     "lengthof(${1:invalue})$0",
		Signature:      "lengthof(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __lengthof__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "sizeof(...)",
		InsertText:     "sizeof(${1:invalue})$0",
		Signature:      "sizeof(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __sizeof__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "ispresent(...)",
		InsertText:     "ispresent(${1:invalue})$0",
		Signature:      "ispresent(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __ispresent__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "ischosen(...)",
		InsertText:     "ischosen(${1:invalue})$0",
		Signature:      "ischosen(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __ischosen__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "isvalue(...)",
		InsertText:     "isvalue(${1:invalue})$0",
		Signature:      "isvalue(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __isvalue__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "isbound(...)",
		InsertText:     "isbound(${1:invalue})$0",
		Signature:      "isbound(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __isbound__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "istemplatekind(...)",
		InsertText:     "istemplatekind(${1:invalue})$0",
		Signature:      "istemplatekind(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __istemplatekind__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:          "regexp(...)",
		InsertText:     "regexp(${1:invalue})$0",
		Signature:      "regexp(in integer invalue) return charstring",
		Documentation:  "## (TTCN-3)\nThe __regexp__ function ",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "substr(...)",
		InsertText: "substr(${1:inpar}, ${2:index}, ${3:count})$0",
		Signature:  `substr(in template (present) any inpar, in integer index, in integer count) return input_string_or_sequence_type`,
		Documentation: `## (TTCN-3)
This function returns a substring or subsequence from a value that is of a binary string type (__bitstring__,
__hexstring__, __octetstring__), a character string type (__charstring__, __universal charstring__), or a sequence
type (__record of__, __set of__ or array). If _inpar_ is a literal (i.e. type is not explicitly given) the corresponding type
shall be retrieved from the value contents. The type of the substring or subsequence returned is the root type of the input
parameter. The starting point of substring or subsequence to return is defined by the second parameter (_index_).
Indexing starts from zero. The third input parameter (_count_) defines the length of the substring or subsequence to be
returned. The units of length for string types are as defined in table 4 of the TTCN-3 core language specification. For sequence types, the
unit of length is element.

Please note that the root types of arrays is __record of__, therefore if _inpar_ is an array the returned type
is __record of__. This, in some cases, may lead to different indexing in _inpar_ and in the returned value.

When used on templates of character string types, specific values and patterns that contain literal characters and the
following metacharacters: "?", "*" are allowed in _inpar_ and the function shall return the character representation of
the matching mechanisms. When _inpar_ is a template of binary string or sequence type or is an array, only the specific
value combined templates whose elements are specific values, and AnyElement matching mechanisms or combined
templates are allowed and the substring or subsequence to be returned shall not contain AnyElement or combined
template.

In addition to the general error causes in clause 16.1.2, error causes are:

* _index_ is less than zero;
* _count_ is less than zero;
* _index_ + _count___ is greater than __lengthof__(_inpar_);
* _inpar_ is a template of a character string type and contains a matching mechanism other than a specific value
or pattern; or if the pattern contains other metacharacters than "?", "*";
* _inpar_ is a template of a binary string or sequence type or array and it contains other matching mechanism as
specific value or combined template; or if the elements of combined template are any other matching
mechanisms than specific values, and AnyElement or combined templates;
* _inpar_ is a template of a binary string or sequence type or array and the substring or subsequence to be
returned contains the AnyElement matching mechanism or combined templates;
* the template passed to the _inpar_ parameter is not of type bitstring, hexstring, octetstring,
charstring, universal charstring, __record of__, __set of__, or array.

Examples:

	substr('00100110'B, 3, 4) // returns '0011'B
	substr('ABCDEF'H, 2, 3) // returns 'CDE'H
	substr('01AB23CD'O, 1, 2) // returns 'AB23'O
	substr("My name is JJ", 11, 2) // returns "JJ"
	substr({ 4, 5, 6 }, 1, 2) // returns {5, 6}`,
		NrOfParameters: 3,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "replace(...)",
		InsertText: "replace(${1:inpar}, (${2:index}, (${3:len}, (${4:repl})$0",
		Signature:  `replace(in any inpar, in integer index, in integer len, in any repl) return any_string_or_sequence type`,
		Documentation: `## (TTCN-3)
This function replaces the substring or subsequence of value _inpar_ at index _index_ of length _len_ with the string or
sequence value _repl_ and returns the resulting string or sequence. _inpar_ shall not be modified. If _len_ is 0 the string
or sequence _repl_ is inserted. If _index_ is 0, _repl_ is inserted at the beginning of _inpar_. If _index_ is
__lengthof(_inpar_), _repl_ is inserted at the end of _inpar_. If _inpar_ is a literal (i.e. type is not explicitly given) the
corresponding type shall be retrieved from the value contents. _inpar_ and _repl_, and the returned string or sequence
shall be of the same root type. The function replace can be applied to _bitstring_, _hexstring_, _octetstring_, or
any character string, __record of__, __set of__, or arrays. Note that indexing in strings starts from zero.

Please note that the root types of arrays is __record of__, therefore if _inpar_ or _repl_ or both are an
array, the returned type is __record of__. This, in some cases, may lead to different indexing in _inpar_
and/or _repl_ and in the returned value.

In addition to the general error causes in clause 16.1.2, error causes are:
* _inpar_ or _repl_ are not of string, __record of__, __set of__, or array type;
* _inpar_ and _repl_ are of different root type;
* _index_ is less than 0 or greater than __lengthof(_inpar_);
* _len_ is less than 0 or greater than __lengthof(_inpar_);
* _index_+_len_ is greater than __lengthof(_inpar_).

Examples:

	replace ('00000110'B, 1, 3, '111'B) // returns '01110110'B
	replace ('ABCDEF'H, 0, 2, '123'H) // returns '123CDEF'H
	replace ('01AB23CD'O, 2, 1, 'FF96'O) // returns '01ABFF96CD'O
	replace ("My name is JJ", 11, 1, "xx") // returns "My name is xxJ"
	replace ("My name is JJ", 11, 0, "xx") // returns "My name is xxJJ"
	replace ("My name is JJ", 2, 2, "x") // returns "Myxame is JJ",
	replace ("My name is JJ", 12, 2, "xx") // produces test case error
	replace ("My name is JJ", 13, 2, "xx") // produces test case error
	replace ("My name is JJ", 13, 0, "xx") // returns "My name is JJxx"`,
		NrOfParameters: 4,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "encvalue(...)",
		InsertText: "encvalue(${1:inpar}, ${2:encoding_info}, ${3:dynamic_encoding})$0",
		Signature:  `encvalue(in template (value) any inpar, in universal charstring encoding_info := "", in universal charstring dynamic_encoding := "") return bitstring`,
		Documentation: `## (TTCN-3)
The __encvalue__ function encodes a value or template into a bitstring. When the actual parameter that is passed to
_inpar_ is a template, it shall resolve to a specific value (the same restrictions apply as for the argument of the send
statement). The returned bitstring represents the encoded value of _inpar_, however, the TTCN-3 test system need not
make any check on its correctness. The optional _encoding_info_ parameter is used for passing additional encoding
information to the codec and, if it is omitted, no additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of encode attribute of the _inpar_ value
for this single __encvalue__ call. The rules for dynamic selection of the encode attribute are described in clause 27.9 of the TTCN-3 core language specification.

In addition to the general error causes in clause 16.1.2, error causes are:

* Encoding fails due to a runtime system problem (i.e. no encoding function exists for the actual type of
_inpar_).`,
		NrOfParameters: 3,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "decvalue(...)",
		InsertText: "decvalue(${1:encoded_value}, ${2:decoded_value}, ${3:decoding_info}, ${4:dynamic_encoding})$0",
		Signature:  `decvalue(inout bitstring encoded_value, out any decoded_value, in universal charstring decoding_info := "", in universal charstring dynamic_encoding := "") return integer`,
		Documentation: `## (TTCN-3)
The __decvalue__ function decodes a bitstring into a value. The test system shall suppose that the bitstring
_encoded_value_ represents an encoded instance of the actual type of _decoded_value_. The optional
_decoding_info_ parameter is used for passing additional decoding information to the codec and, if it is omitted, no
additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of encode attribute of the
_decoded_value_ parameter for this single __decvalue__ call. The rules for dynamic selection of the encode attribute
are described in clause 27.9 of the TTCN-3 core language specification.

If the decoding was successful, then the used bits are removed from the parameter _encoded_value_, the rest is
returned (in the parameter _encoded_value_), and the decoded value is returned in the parameter _decoded_value_.  
If the decoding was unsuccessful, the actual parameters for _encoded_value_ and _decoded_value_ are not
changed. The function shall return an integer value to indicate success or failure of the decoding below:

* The return value 0 indicates that decoding was successful.
* The return value 1 indicates an unspecified cause of decoding failure. This value is also returned if the
_encoded_value_ parameter contains an unitialized value.
* The return value 2 indicates that decoding could not be completed as _encoded_value_ did not contain
enough bits.`,
		NrOfParameters: 4,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "encvalue_unichar(...)",
		InsertText: "encvalue_unichar(${1:inpar}, ${2:string_serialization}, ${3:encoding_info}, ${4:dynamic_encoding})$0",
		Signature:  `encvalue_unichar(in template (value) any inpar, in charstring string_serialization := "UTF-8", in universal charstring encoding_info := "", in universal charstring dynamic_encoding := "") return universal charstring`,
		Documentation: `## (TTCN-3)
The __encvalue_unichar__ function encodes a value or template into a universal charstring. When the actual
parameter that is passed to _inpar_ is a template, it shall resolve to a specific value (the same restrictions apply as for
the argument of the send statement). The returned universal charstring represents the encoded value of _inpar_,
however, the TTCN-3 test system need not make any check on its correctness. If the optional _string_serialization_
parameter is omitted, the default value "UTF-8" is used. The optional _encoding_info_ parameter is used for passing
additional encoding information to the codec and, if it is omitted, no additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of encode attribute of the _inpar_ value
for this single __encvalue_unichar__ call. The rules for dynamic selection of the encode attribute are described in
clause 27.9 of the TTCN-3 core language specification.

The following values (see ISO/IEC 10646 [2]) are allowed as _string_serialization_ actual parameters (for the description
of the UCS encoding scheme see clause 27.5):
* "UTF-8"
* "UTF-16"
* "UTF-16LE"
* "UTF-16BE"
* "UTF-32"
* "UTF-32LE"
* "UTF-32BE"

The serialized bitstring shall not include the optional signature (see clause 10 of ISO/IEC 10646 [2], also known as byte
order mark).

In case of "UTF-16" and "UTF-32" big-endian ordering shall be used (as described in clauses 10.4 and 10.7 of
ISO/IEC 10646 [2]).

The specific semantics of this function are explained by the following TTCN-3 definition:

	function encvalue_unichar(in template(value) any inpar,
	            in charstring enc
	            in universal charstring encoding_info := "",
	            in universal charstring dynamic_encoding := "") return universal charstring {
		return oct2unichar(bit2oct(encvalue(inpar, encoding_info, dynamic_encoding)), enc);
	}`,
		NrOfParameters: 4,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "decvalue_unichar(...)",
		InsertText: "decvalue_unichar(${1:encoded_value}, ${2:decoded_value}, ${3:string_serialization}, ${4:decoding_info}, ${5:dynamic_encoding},)$0",
		Signature:  `decvalue_unichar(inout universal charstring encoded_value, out any decoded_value, in charstring string_serialization:= "UTF-8", in universal charstring decoding_info := "", in universal charstring dynamic_encoding := "") return integer`,
		Documentation: `## (TTCN-3)
The __decvalue_unichar__ function decodes (part of) a universal charstring into a value. The test system shall
suppose that a prefix of the universal charstring _encoded_value_ represents an encoded instance of the actual type of
_decoded_value_. The optional _decoding_info_ parameter is used for passing additional decoding information to
the codec and, if it is omitted, no additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of encode attribute of the
_decoded_value_ parameter for this single __decvalue_unichar__ call. The rules for dynamic selection of the
encode attribute are described in clause 27.9.

If the decoding was successful, then the characters used for decoding are removed from the parameter
_encoded_value_, the rest is returned (in the parameter _encoded_value_), and the decoded value is returned in the
parameter _decoded_value_. If the decoding was unsuccessful, the actual parameters for _encoded_value_ and
_decoded_value_ are not changed. The function shall return an integer value to indicate success or failure of the
decoding below:

* The return value 0 indicates that decoding was successful.
* The return value 1 indicates an unspecified cause of decoding failure. This value is also returned if the
_encoded_value_ parameter contains an unitialized value.
* The return value 2 indicates that decoding could not be completed as _encoded_value_ did not contain
enough bits.

If the optional _string_serialization_ parameter is omitted, the default value "UTF-8" is used.

The following values (see ISO/IEC 10646 [2]) are allowed as _string_serialization_ actual parameters (for the description
of the UCS encoding scheme see clause 27.5 of TTCN-3 core language specification):

* "UTF-8"
* "UTF-16"
* "UTF-16LE"
* "UTF-16BE"
* "UTF-32"
* "UTF-32LE"
* "UTF-32BE"

The serialized bitstring shall not include the optional signature (see clause 10 of ISO/IEC 10646 [2], also known as byte
order mark).

In case of "UTF-16" and "UTF-32" big-endian ordering shall be used (as described in clauses 10.4 and 10.7 of
ISO/IEC 10646 [2]).  
The semantics of the function can be explained by the following TTCN-3 function:

	function decvalue_unichar (
	            inout universal charstring encoded_value,
	            out any decoded_value,
	            in charstring string_encoding := "UTF-8",
	            in universal charstring decoding_info := "",
	            in universal charstring dynamic_encoding := "") return integer {
		var bitstring v_str = oct2bit(unichar2oct(encoded_value, string_encoding));
		var integer v_result := decvalue(v_str, decoded_value, decoding_info, dynamic_encoding);
		if (v_result == 0) { // success
			encoded_value := oct2unichar(bit2oct(v_str), string_encoding);
		}
		return v_result;
	}`,
		NrOfParameters: 5,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "encvalue_o(...)",
		InsertText: "encvalue_o(${1:inpar}, ${2:encoding_info}, ${3:dynamic_encoding}, ${4:bit_length})$0",
		Signature:  `encvalue_o(in template (value) any inpar, in universal charstring encoding_info := "", in universal charstring dynamic_encoding := "", out integer bit_length) return octetstring`,
		Documentation: `## (TTCN-3)
The __encvalue_o__ function encodes a value or template into an octetstring. When the actual parameter that is passed
to _inpar_ is a template, it shall resolve to a specific value (the same restrictions apply as for the argument of the send
statement). The returned octetstring represents the encoded value of _inpar_, however, the TTCN-3 test system need not
make any check on its correctness. In case the encoded message is not octet-based and has a bit length not divisable by
8, the encoded message will be left-aligned in the returned octetstring and the least significant (8 - (bit length mod 8))
bits in the least significant octet will be 0. The bit length can be assigned to a variable by usage of the formal out
parameter _bit_length_. The optional _encoding_info_ parameter is used for passing additional encoding
information to the codec and, if it is omitted, no additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of encode attribute of the _inpar_ value
for this single __encvalue_o__ call. The rules for dynamic selection of the encode attribute are described in clause 27.9.

In addition to the general error causes in clause 16.1.2 of the TTCN-3 core language specification, error causes are:
* Encoding fails due to a runtime system problem (i.e. no encoding function exists for the actual type of
_inpar_).`,
		NrOfParameters: 4,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "decvalue_o(...)",
		InsertText: "decvalue_o(${1:encoded_value}, ${2:decoded_value}, ${3:decoding_info}, ${4:dynamic_encoding})$0",
		Signature: `decvalue_o(inout octetstring encoded_value,
out any decoded_value,
in universal charstring decoding_info := "",
in universal charstring dynamic_encoding := "") return integer`,
		Documentation: `## (TTCN-3)
The __decvalue_o__ function decodes an octetstring into a value. The test system shall suppose that the octetstring
_encoded_value_ represents an encoded instance of the actual type of _decoded_value_. The optional
_decoding_info_ parameter is used for passing additional decoding information to the codec and, if it is omitted, no
additional information is sent to the codec.

The optional _dynamic_encoding_ parameter is used for dynamic selection of __encode__ attribute of the
decoded_value parameter for this single __decvalue_o__ call. The rules for dynamic selection of the __encode__
attribute are described in clause 27.9 of the TTCN-3 core language specification.

If the decoding was successful, then the used octets are removed from the parameter _encoded_value_, the rest is
returned (in the parameter _encoded_value_), and the decoded value is returned in the parameter _decoded_value_.
If the decoding was unsuccessful, the actual parameters for _encoded_value_ and _decoded_value_ are not
changed. The function shall return an integer value to indicate success or failure of the decoding below:

* The return value 0 indicates that decoding was successful.
* The return value 1 indicates an unspecified cause of decoding failure. This value is also returned if the
_encoded_value_ parameter contains an unitialized value.
* The return value 2 indicates that decoding could not be completed as _encoded_value_ did not contain
enough octets.
`,
		NrOfParameters: 4,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "get_stringencoding(...)",
		InsertText: "get_stringencoding(${1:encoded_value})$0",
		Signature:  "get_stringencoding(in octettstring encoded_value) return octettstring",
		Documentation: `## (TTCN-3)
The __get_stringencoding__ function analyses the encoded_value and returns the UCS encoding scheme according to
clause 10 of ISO/IEC 10646 [2] (see also clause 27.5 of the TTCN-3 core language specification). The identified encoding scheme, or the
value "<unknown>", if the type of encoding cannot be determined unanimously, shall be returned as a character string.

The initial octet sequence (also known as byte order mark, BOM), when present, allows identifying the
encoding scheme unanimously. When it is not present, other symptoms may be used to identify the
encoding scheme unanimously; for example, only UTF-8 may have odd number of octets and bit
distribution according to table 2 of clause 9.1 of ISO/IEC 10646 [2].

Example:

    match ( get_stringencoding('6869C3BA7A'O),charstring:"UTF-8") // true
    //(the octetstring contains the UTF-8 encoding of the character sequence "hiúz")
`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "remove_bom(...)",
		InsertText: "remove_bom(${1:encoded_value})$0",
		Signature:  "remove_bom(in octettstring encoded_value) return octettstring",
		Documentation: `## (TTCN-3)
The __remove_bom__ function removes the optional FEFF ZERO WIDTH NO-BREAK SPACE sequence that may be
present at the beginning of a stream of serialized (encoded) universal character strings to indicate the order of the octets
within the encoding form, as defined in clause 10 of ISO/IEC 10646 [2]. If no FEFF ZERO WIDTH NO-BREAK
SPACE sequence present in the _encoded_value_ parameter, the function shall return the value of the parameter
without change.`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "rnd(...)",
		InsertText: "rnd(${1:seed})$0",
		Signature:  "rnd([in float seed]) return float",
		Documentation: `## (TTCN-3)
The __rnd__ function returns a (pseudo) random number less than 1 but greater or equal to 0. The random number
generator is initialized per test component and for the control part by means of an optional seed value (a numerical float
value). If no new seed is provided, the last generated number will be used as seed for the next random number. Without
a previous initialization a value calculated from the system time will be used as seed value when __rnd__ is used the first
time in a test component or the control part.

Each time the __rnd__ function is initialized with the same seed value, it shall repeat the same sequence of random
numbers.

For the purpose of keeping parallel testing deterministic, each test component, as well as the control part
has its own random seed. This allows for better reproducibility of test executions. Thus, the __rnd__ function
will always use the seed of the component or control part which calls it.

To produce a random integers in a given range, the following formula can be used:

    float2int(int2float(upperbound - lowerbound +1)*rnd()) + lowerbound
    // Here, upperbound and lowerbound denote highest and lowest number in range.

`,
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "testcasename()",
		InsertText: "testcasename()$0",
		Signature:  "testcasename() return charstring",
		Documentation: `## (TTCN-3)
The __testcasename__ function shall return the unqualified name of the actually executing test case.

When the function __testcasename__ is called if the control part is being executed but no testcase, it shall return the
empty string.`,

		NrOfParameters: 0,
		TextFormat:     protocol.SnippetTextFormat},
	{
		Label:      "hostid(...)",
		InsertText: "hostid($1)$0",
		Signature:  "hostid(in charstring idkind := \"Ipv4orIPv6\") return charstring",
		Documentation: "## (TTCN-3)\nThe __hostid__ function shall return the host id of the test component or module control executing the hostid function	in form of a character string. The in parameter idkind allows to specify the expected id format to be returned.	Predefined _idkind_ values are:\n* \"Ipv4orIPv6\": The contents of the returned character string is an Ipv4 address. If no Ipv4 address, but an	Ipv6 address is available, a character string representation of the Ipv6 address is returned.\n* \"Ipv4\": The contents of the returned character string shall be an Ipv4 address.\n* \"Ipv6\": The contents of the returned characterstring shall be an Ipv6 address.",
		NrOfParameters: 1,
		TextFormat:     protocol.SnippetTextFormat}}
