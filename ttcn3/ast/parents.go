package ast

import (
	"reflect"
)

func Parents(tgt, root Node) []Node {
	var (
		path  []Node
		visit func(n Node)
	)

	pos := tgt.Pos()
	visit = func(n Node) {
		if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() || n == nil {
			return
		}

		if inside := n.Pos() <= pos && pos < n.End(); inside {
			if n == tgt {
				return
			}
			path = append(path, n)
			for _, child := range Children(n) {
				visit(child)
			}
		}

	}
	visit(root)

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}

	return path
}
