package loader

import (
	"runtime"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

var ioLimit = make(chan bool, runtime.NumCPU()*50)

func parseFiles(fset *loc.FileSet, files []string) ([]*ast.Module, error) {
	var wg sync.WaitGroup
	n := len(files)
	mods := make([][]*ast.Module, n)
	errors := make([]error, n)
	for i, file := range files {
		wg.Add(1)
		go func(i int, file string) {
			ioLimit <- true // wait
			defer func() {
				wg.Done()
				<-ioLimit // signal
			}()
			mods[i], errors[i] = parser.ParseModules(fset, file, nil, 0, nil)
		}(i, file)
	}
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	var list []*ast.Module
	for _, m := range mods {
		list = append(list, m...)
	}
	return list, nil
}
