// Package fs provides a primitive virtual file system.
package fs

var store = Store{}

// Open a file.
//
// path can be any identifier, URL, ...
func Open(path string) *File {
	return store.Open(path)
}

// PathSlice returns a string slice of the File objects passed as argument.
func PathSlice(files ...*File) []string {
	ret := make([]string, 0, len(files))
	for i := range files {
		if files[i] != nil {
			ret = append(ret, files[i].String())
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}
