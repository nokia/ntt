package lsp

import (
	"encoding/json"
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/span"
)

func location(pos loc.Position) protocol.Location {
	return protocol.Location{
		URI: protocol.URIFromSpanURI(span.URIFromPath(pos.Filename)),
		Range: protocol.Range{
			Start: position(pos.Line, pos.Column),
			End:   position(pos.Line, pos.Column),
		},
	}
}

func position(line, column int) protocol.Position {
	return protocol.Position{
		Line:      float64(line - 1),
		Character: float64(column - 1),
	}
}

func marshalRaw(vs ...interface{}) ([]json.RawMessage, error) {
	var ret []json.RawMessage
	for _, v := range vs {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func unmarshalRaw(bs []json.RawMessage, vs ...interface{}) error {
	if len(bs) != len(vs) {
		return fmt.Errorf("unexpected number of arguments")
	}
	for i, b := range bs {
		if err := json.Unmarshal(b, &vs[i]); err != nil {
			return err
		}
	}
	return nil
}
