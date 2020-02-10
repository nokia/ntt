package doc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTag(t *testing.T, txt string, expect []string) {
	f := FindTag(txt)
	if len(f) == 2 {
		assert.Equal(t, expect[0], f[0])
		assert.Equal(t, expect[1], f[1])
	}
}

func TestTags(t *testing.T) {
	testTag(t, `@foo: bar`, []string{`@foo`, `bar`})
	testTag(t, `@foo1: bar1
	            @foo2: bar2`, []string{`@foo1`, `bar1`})
	testTag(t, `// @foo	bar`, []string{`@foo`, `bar`})
	testTag(t, `// @foo		@bar	`, []string{`@foo`, `@bar`})
	testTag(t, `// *** @foo	@bar ***`, []string{`@foo`, `@bar`})
	testTag(t, `// *** @wip23	***`, []string{`@wip23`, ``})
	testTag(t, `** @verdict  @foo:bar`, []string{`@verdict`, `@foo:bar`})
	testTag(t, `** @verdict:@foo:bar`, []string{`@verdict`, `@foo:bar`})
}

func TestFindAllTags(t *testing.T) {
	expect := [][]string{
		{`@foo`, `bar`},
		{`@foo`, `bar`},
		{`@foo`, `@bar`},
		{`@foo`, `@bar`},
		{`@wip23`, ``},
		{`@verdict`, `@foo:bar`},
		{`@verdict`, `@foo:bar`},
		{`@one`, ``},
		{`@two`, ``},
	}

	actual := FindAllTags(`/*
		@foo: bar
		// @foo	bar
		// @foo		@bar	
		// *** @foo	@bar ***
		// *** @wip23	***
		** @verdict  @foo:bar
		** @verdict:@foo:bar
		* @one
		* @two
		*/`)

	assert.Equal(t, len(expect), len(actual))

	for i, _ := range actual {
		assert.Equal(t, expect[i], actual[i])
	}
}
