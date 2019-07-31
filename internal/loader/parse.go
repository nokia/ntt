package loader

import (
	"runtime"
	"sync"

	"github.com/nokia/ntt/internal/ttcn3/syntax"
)

var ioLimit = make(chan bool, runtime.NumCPU()*50)

func parseFiles(fset *syntax.FileSet, files []string) ([]*syntax.Module, error) {
	var wg sync.WaitGroup
	n := len(files)
	mods := make([][]*syntax.Module, n)
	errors := make([]error, n)
	for i, file := range files {
		wg.Add(1)
		go func(i int, file string) {
			ioLimit <- true // wait
			defer func() {
				wg.Done()
				<-ioLimit // signal
			}()
			mods[i], errors[i] = syntax.ParseModules(fset, file, nil, 0, nil)
		}(i, file)
	}
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	var list []*syntax.Module
	for _, m := range mods {
		list = append(list, m...)
	}
	return list, nil
}
