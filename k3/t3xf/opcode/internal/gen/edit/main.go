// This program helps to modify opcodes.yml programmatically.

package main

import (
	"fmt"
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

	//removeContext(opcodes)
	addContext(opcodes)

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

func addContext(opcodes map[string]*opcode.Descriptor) {

	contexts := map[string][]string{
		"start":  {"port_op", "timer_op"},
		"startc": {"component_op"},
		"stopac": {"component_op"},
		"alive":  {"component_op"},
		"alive1": {"component_op", "any"},
		"alivea": {
			"component_op",
			"all",
		},
		"done": {
			"component_op",
		},
		"done1": {
			"component_op",
			"any",
		},
		"donea": {
			"component_op",
			"all",
		},
		"running": {
			"timer_op",
			"component_op",
		},
		"running1c": {
			"component_op",
			"timer_op",
			"any",
		},
		"runningac": {
			"component_op",
			"timer_op",
			"all",
		},
		"killed": {"component_op"},
		"killed1": {
			"component_op",
			"any",
		},
		"killeda": {
			"component_op",
			"all",
		},
		"create":   {"component_new"},
		"createa":  {"component_new"},
		"createan": {"component_new"},
		"createn":  {"component_new"},
		"kill":     {"component_op"},
		"killa": {
			"component_op",
			"all",
		},
		"getverdict": {"component_op"},
		"setverdict": {"component_op"},
		"activate":   {"component_op"},
		"deactivate": {"component_op"},
		"deactivatea": {
			"component_op",
			"all",
		},
		"startap": {
			"port_op",
			"all",
		},
		"stopap": {
			"port_op",
			"all",
		},
		"clear": {
			"port_op",
		},
		"cleara": {
			"port_op",
			"all",
		},
		"checkstate": {
			"port_op",
		},
		"checkstateal": {
			"port_op",
			"all",
		},
		"checkstatean": {
			"port_op",
			"any",
		},
		"connect": {
			"port_op",
		},
		"disconnect": {
			"port_op",
		},
		"disconnecta": {
			"port_op",
			"all",
		},
		"disconnectaa": {
			"port_op",
			"component_op",
			"all",
		},
		"map":   {"port_op"},
		"unmap": {"port_op"},
		"unmapa": {
			"port_op",
			"all",
		},
		"unmapaa": {
			"port_op",
			"component_op",
			"all",
		},
		"halt": {
			"port_op",
		},
		"halta": {
			"port_op",
			"all",
		},
		"send":  {"port_op"},
		"send1": {"port_op"},
		"senda": {"port_op"},
		"sendn": {"port_op"},
		"check": {"port_op"},
		"check1": {
			"port_op",
			"any",
		},
		"receive":  {"port_op"},
		"receive1": {"port_op"},
		"receivec": {"port_op"},
		"receivec1": {
			"port_op",
			"any",
		},
		"trigger": {"port_op"},
		"trigger1": {
			"port_op",
			"any",
		},
		"startd": {"timer_op"},
		"stop": {
			"timer_op",
			"component_op",
			"port_op",
		},
		"stopat": {
			"timer_op",
			"all",
		},
		"read": {"timer_op"},
		"running1t": {
			"timer_op",
			"any",
		},
		"timeout": {"timer_op"},
		"timeout1": {
			"timer_op",
			"any",
		},
		"system":       {"testcase_op"},
		"testcasename": {"testcase_op"},
		"mtc":          {"testcase_op"},
		"self":         {"testcase_op"},
		"alt":          {"eval"},
		"interleave":   {"eval"},
		"else":         {"eval"},
		"step":         {"eval"},
		"specplc":      {"component_op"},
		"exec":         {"vm"},
		"value":        {"port_op"},
		"execute":      {"testcase_new"},
		"executel":     {"testcase_new"},
		"ifpresent":    {"value"},
		"length":       {"value"},
		"allfrom":      {"value", "all"},
		"allfromp":     {"value", "all"},
		"valueof":      {"value"},
		"complement":   {"value"},
		"superset":     {"value"},
		"permutation":  {"value"},
		"subset":       {"value"},
		"action":       {"func"},
		"log":          {"func"},
		"now":          {"func"},
		"wait":         {"func"},
		"decvalue":     {"func"},
		"encvalue":     {"func"},
		"isbound":      {"func"},
		"ischosen":     {"func"},
		"ispresent":    {"func"},
		"isvalue":      {"func"},
		"lengthof":     {"func"},
		"sizeof":       {"func"},
		"match":        {"func"},
		"smatch":       {"func"},
		"regexp":       {"func"},
		"replace":      {"func"},
		"substr":       {"func"},
		"rnd":          {"func"},
		"bit2hex":      {"func"},
		"bit2int":      {"func"},
		"bit2oct":      {"func"},
		"bit2str":      {"func"},
		"char2int":     {"func"},
		"char2oct":     {"func"},
		"enum2int":     {"func"},
		"float2int":    {"func"},
		"hex2bit":      {"func"},
		"hex2int":      {"func"},
		"hex2oct":      {"func"},
		"hex2str":      {"func"},
		"int2bit":      {"func"},
		"int2char":     {"func"},
		"int2enum":     {"func"},
		"int2float":    {"func"},
		"int2hex":      {"func"},
		"int2oct":      {"func"},
		"int2str":      {"func"},
		"oct2bit":      {"func"},
		"oct2chr":      {"func"},
		"oct2hex":      {"func"},
		"oct2int":      {"func"},
		"oct2str":      {"func"},
		"str2float":    {"func"},
		"str2hex":      {"func"},
		"str2int":      {"func"},
		"str2oct":      {"func"},
		"val2str":      {"func"},
		"clone":        {"func"},
		"neg":          {"func"},
		"mul":          {"func"},
		"div":          {"func"},
		"rem":          {"func"},
		"mod":          {"func"},
		"add":          {"func"},
		"sub":          {"func"},
		"cat":          {"func"},
		"eq":           {"func"},
		"ge":           {"func"},
		"gt":           {"func"},
		"le":           {"func"},
		"lt":           {"func"},
		"ne":           {"func"},
		"rol":          {"func"},
		"ror":          {"func"},
		"shl":          {"func"},
		"shr":          {"func"},
		"xor":          {"func"},
		"not":          {"func"},
		"or":           {"func"},
		"and":          {"func"},
		"scan":         {"vm"},
		"block":        {"vm"},
		"nop":          {"vm"},
		"drop":         {"vm"},
		"frozen_ref":   {"vm"},
		"ref":          {"vm"},
		"assign":       {"vm"},
		"assignd":      {"vm"},
		"return":       {"vm", "control_flow"},
		"dowhile":      {"vm", "control_flow"},
		"for":          {"vm", "control_flow"},
		"if":           {"vm", "control_flow"},
		"ifelse":       {"vm", "control_flow"},
		"while":        {"vm", "control_flow"},
		"goto":         {"vm", "control_flow"},
		"natlong": {
			"literal",
			"multibyte",
		},
		"ieee754dp": {
			"literal",
			"multibyte",
		},
		"istr": {
			"literal",
			"multibyte",
		},
		"fstr": {
			"literal",
			"multibyte",
		},
		"name": {
			"literal",
			"multibyte",
		},
		"utf8": {
			"literal",
			"multibyte",
		},
		"octets": {
			"literal",
			"multibyte",
		},
		"nibbles": {
			"literal",
			"multibyte",
		},
		"bits": {
			"literal",
			"multibyte",
		},
		"any":         {"literal"},
		"anyn":        {"literal"},
		"skip":        {"literal"},
		"mark":        {"literal"},
		"null":        {"literal"},
		"omit":        {"literal"},
		"error":       {"literal"},
		"fail":        {"literal"},
		"inconc":      {"literal"},
		"pass":        {"literal"},
		"none":        {"literal"},
		"false":       {"literal"},
		"true":        {"literal"},
		"infinityn":   {"literal"},
		"infinityp":   {"literal"},
		"address":     {"type_ref"},
		"bitstring":   {"type_ref"},
		"boolean":     {"type_ref"},
		"charstring":  {"type_ref"},
		"charstringu": {"type_ref"},
		"default":     {"type_ref"},
		"float":       {"type_ref"},
		"hexstring":   {"type_ref"},
		"integer":     {"type_ref"},
		"octetstring": {"type_ref"},
		"verdicttype": {"type_ref"},
		"timer":       {"type_ref"},
		"closuretype": {"type_ref"},
		"component":   {"type_new"},
		"componentx":  {"type_new"},
		"enumerated":  {"type_new"},
		"portm":       {"type_new"},
		"portma":      {"type_new"},
		"record":      {"type_new"},
		"set":         {"type_new"},
		"recordof":    {"type_new"},
		"setof":       {"type_new"},
		"subtype":     {"type_new"},
		"union":       {"type_new"},
		"array":       {"type_new"},
		"type":        {"type_def"},
		"typew":       {"type_def"},
		"source":      {"vm"},
		"module":      {"vm"},
		"const":       {"decl"},
		"constw":      {"decl"},
		"var":         {"decl"},
		"vardup":      {"decl"},
		"mpar":        {"decl"},
		"mpard":       {"decl"},
		"template":    {"beha_def"},
		"control":     {"beha_def"},
		"altstep":     {"beha_def"},
		"altstepb":    {"beha_def"},
		"altstepbw":   {"beha_def"},
		"altstepw":    {"beha_def"},
		"function":    {"beha_def"},
		"functionb":   {"beha_def"},
		"functionv":   {"beha_def"},
		"functionvb":  {"beha_def"},
		"functionxv":  {"beha_def"},
		"functionxvw": {"beha_def"},
		"functionxw":  {"beha_def"},
		"testcase":    {"beha_def"},
		"testcases":   {"beha_def"},
		"version":     {"beha_def"},
		"term":        {"type_new"},
		"field":       {"type_new"},
		"fieldo":      {"type_new"},
		"ifield":      {"type_new"},
		"in":          {"type_new"},
		"inout":       {"type_new"},
		"out":         {"type_new"},
		"encode":      {"type_new"},
		"encodeo":     {"type_new"},
		"extension":   {"type_new"},
		"extensiono":  {"type_new"},
		"variant":     {"type_new"},
		"varianto":    {"type_new"},
		"permito":     {"vm"},
		"permitp":     {"vm"},
		"permitt":     {"vm"},
		"at_default":  {"vm"},
		"def":         {"vm"},
		"idef":        {"vm"},
		"iget":        {"vm"},
		"get":         {"vm"},
		"apply":       {"vm"},
		"collect":     {"vm"},
		"line":        {"vm"},
		"vlist":       {"vm"},
		"closure":     {"vm"},
		"pattern":     {"vm"},
		"range":       {"vm"},
		"unmapfromto": {"value"},
		"mapt":        {"type_new"},
		"decmatch":    {"value"},
		"to":          {"func"},
		"from":        {"func"},
	}
	for k, v := range contexts {
		if opcodes[k] != nil {
			opcodes[k].Context = v
		} else {
			fmt.Printf("Opcode %s not found\n", k)
		}
	}
}
