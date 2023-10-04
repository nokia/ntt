package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/stretchr/testify/assert"
)

func TestPlainTextHoverForFunction(t *testing.T) {
	actual := testHover(t, `
        module Test {
            function myfunc(integer x)
                runs on Component
                system System
                return integer
            {
                return x;
            }
        }`,
		protocol.Position{Line: 2, Character: 25},
		&lsp.PlainTextHover{})

	expected :=
		"function myfunc(integer x)\n" +
			"  runs on Component\n" +
			"  system System\n" +
			"  return integer\n"

	assert.Equal(t, expected, actual.Contents.Value)
}

func TestMarkdownHoverForFunction(t *testing.T) {
	actual := testHover(t, `
        module Test {
            function myfunc(integer x)
                runs on Component
                system System
                return integer
            {
                return x;
            }
        }`,
		protocol.Position{Line: 2, Character: 25},
		&lsp.MarkdownHover{})

	expected :=
		"```typescript\n" +
			"function myfunc(integer x)\n" +
			"  runs on Component\n" +
			"  system System\n" +
			"  return integer\n" +
			"```\n" +
			" - - -\n" +
			"\n" +
			" - - -\n"

	assert.Equal(t, expected, actual.Contents.Value)
}

func testHover(t *testing.T, text string, position protocol.Position, capability lsp.HoverContentProvider) *protocol.Hover {
	t.Helper()

	file := fmt.Sprintf("%s.ttcn3", t.Name())
	fs.SetContent(file, []byte(text))

	params := protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			Position: position,
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(file),
			},
		},
	}

	hover, err := lsp.ProcessHover(&params, &ttcn3.DB{}, capability)
	assert.Equal(t, err, nil)

	return hover
}
