package dump

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	indent = 0

	Command = &cobra.Command{
		Use:    "dump",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			suite, err := ntt.NewFromArgs(args...)
			if err != nil {
				return err
			}

			files, _ := project.Files(suite)
			for _, file := range files {
				info := suite.Parse(file)
				dump(reflect.ValueOf(info.Module), "Root: ")
			}

			return nil
		},
	}
)

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
