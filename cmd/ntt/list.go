package main

import (
	"bufio"
	"fmt"
	"os"

	"ntt/internal/loader"
	"ntt/internal/ttcn3/syntax"

	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list [tests|modules|imports] [files]",
		Short: "List various types of objects",
		Long: `List various types of objects.

Default output shows the testcase names in current directory.`,

		RunE:          list,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	w = bufio.NewWriter(os.Stdout)
)

var printers = map[string]func(string, *syntax.Module, syntax.Node){
	"tc":        printTests,
	"tcs":       printTests,
	"test":      printTests,
	"tests":     printTests,
	"testcase":  printTests,
	"testcases": printTests,
	"mod":       printModules,
	"mods":      printModules,
	"module":    printModules,
	"modules":   printModules,
	"import":    printImports,
	"imports":   printImports,
}

func printModules(file string, m *syntax.Module, n syntax.Node) {
	if Verbose {
		fmt.Fprint(w, file, ": ")
	}
	fmt.Fprintln(w, m.Name.Tok.Lit)
}

func printTests(file string, m *syntax.Module, n syntax.Node) {
	for _, def := range m.Decls {
		switch x := def.Def.(type) {
		case *syntax.GroupDecl:
		case *syntax.ControlPart:
			if Verbose {
				fmt.Fprint(w, file, ": ")
			}
			fmt.Fprintln(w, m.Name.Tok.Lit+".control")

		case *syntax.FuncDecl:
			if x.Kind.Kind != syntax.TESTCASE {
				break
			}
			if Verbose {
				fmt.Fprint(w, file, ": ")
			}
			fmt.Fprintln(w, m.Name.Tok.Lit+"."+x.Name.Tok.Lit)
		}
	}
}

func printImports(file string, m *syntax.Module, n syntax.Node) {
	for _, def := range m.Decls {
		switch x := def.Def.(type) {
		case *syntax.GroupDecl:
		case *syntax.ImportDecl:
			if Verbose {
				fmt.Fprint(w, file, ": ")
			}
			fmt.Fprintln(w, m.Name.Tok.Lit, "<-", x.Module.Tok.Lit)
		}
	}
}

func list(cmd *cobra.Command, args []string) error {
	printer := printTests

	if len(args) > 0 {
		if f, ok := printers[args[0]]; ok {
			printer = f
			args = args[1:]
		}
	}

	var conf loader.Config
	err := conf.FromArgs(args)
	if err != nil {
		return err
	}

	pkg, err := conf.Load()
	if err != nil {
		return err
	}

	for _, m := range pkg.Modules {
		var file string
		if Verbose {
			file = pkg.Fset.File(m.Tok.Pos()).Name()
		}
		printer(file, m, m)
	}
	w.Flush()
	return nil
}
