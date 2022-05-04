package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/printer"
	"github.com/spf13/cobra"
)

var (
	indent = 0

	DumpCommand = &cobra.Command{
		Use:   "dump",
		Short: "Dump TTCN-3 syntax trees",
		RunE: func(cmd *cobra.Command, args []string) error {
			srcs, _ := fs.TTCN3Files(Project.Sources...)
			imports, _ := fs.TTCN3Files(Project.Imports...)
			files := append(srcs, imports...)
			for _, file := range files {
				tree := ttcn3.ParseFile(file)
				if useTTCN3 {
					printer.Print(os.Stdout, tree.FileSet, tree.Root)
				} else if useDot {
					dot(tree.Root)
				} else {
					dump(reflect.ValueOf(tree.Root), "Root: ")
				}
			}

			return nil
		},
	}

	useTTCN3 = false
	useDot   = false
)

func init() {
	DumpCommand.PersistentFlags().BoolVarP(&useTTCN3, "ttcn3", "", false, "formatted TTCN-3 output")
	DumpCommand.PersistentFlags().BoolVarP(&useDot, "dot", "", false, "graphviz output")
}

func dump(v reflect.Value, f string) {
	if !v.IsValid() || v.IsZero() {
		return
	}

	if v.Kind() != reflect.Interface && v.CanInterface() {
		switch x := v.Interface().(type) {
		case ast.Token:
			if x.IsValid() {
				Prefix(x, f)
				fmt.Printf("[35m%s[0m", x.String())
				Suffix(x)
			}
		case ast.Node:
			if x != nil {
				Prefix(x, f)
				fmt.Printf("[0m%s[0m", strings.TrimPrefix(v.Type().String(), "*ast."))
				Suffix(x)
			}
		}
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v2 := v.Elem(); v2.IsValid() {
			dump(v.Elem(), "")
		}

	case reflect.Interface:
		if v2 := v.Elem(); v2.IsValid() {
			dump(v2, f)
		}

	case reflect.Struct:
		indent++
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			tf := v.Type().Field(i)
			dump(f, fmt.Sprintf("%v: ", tf.Name))
		}
		indent--

	case reflect.Slice:
		Prefix(nil, f)
		Suffix(nil)
		indent++
		for i := 0; i < v.Len(); i++ {
			dump(v.Index(i), fmt.Sprintf("%d: ", i))
		}
		indent--
	}

}

func Prefix(n ast.Node, f string) {
	fmt.Printf("%-15s %s[1m%s[0m", Pos(n), Indent(), f)
}

func Suffix(n ast.Node) {
	fmt.Println()
}

func Pos(n ast.Node) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("[%d-%d)", n.Pos()-1, n.End()-1)
}

func Indent() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "[0;1m")
	for i := 0; i < indent; i++ {
		fmt.Fprint(&buf, "[37;0m·   [0m")
	}
	return buf.String()
}
