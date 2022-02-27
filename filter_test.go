package ntt_test

import (
	"strings"
	"testing"

	"github.com/nokia/ntt"
	"github.com/nokia/ntt/ttcn3/doc"
)

func TestBasketMatch(t *testing.T) {
	tests := []struct {
		basket string
		name   string
		tags   []string
		want   bool
	}{
		{basket: "", name: "", want: true},
		{basket: "", name: "", tags: []string{"@wip"}, want: true},
		{basket: "", name: "foo", want: true},
		{basket: "", name: "foo", tags: []string{"@wip"}, want: true},

		{basket: "-r fo.", name: "foobar", want: true},
		{basket: "-r foo -r bar", name: "foobar", want: true},
		{basket: "-r foo|bar", name: "bar", want: true},
		{basket: "-r foo", name: "bar", want: false},

		{basket: "-x fo.", name: "foobar", want: false},
		{basket: "-r foo -x bar", name: "foobar", want: false},
		{basket: "-x foo -r bar", name: "foobar", want: false},
		{basket: "-x foo|bar", name: "bar", want: false},
		{basket: "-x foo", name: "bar", want: true},

		{basket: "-R foo", tags: []string{"@foo bar"}, want: true},
		{basket: "-R foo:bar", tags: []string{"@foo bar"}, want: true},
		{basket: "-R :bar", tags: []string{"@foo bar"}, want: true},
		{basket: "-R :bar", tags: []string{"@foo", "@bar"}, want: false},
		{basket: "-R bar", tags: []string{"@foo bar"}, want: false},
		{basket: "-R bar", tags: []string{"@foo", "@bar"}, want: true},

		{basket: "-R foo|bar", tags: []string{"@foo", "@bar"}, want: true},
		{basket: "-R foo -R bar", tags: []string{"@foo", "@bar"}, want: true},
		{basket: "-R foo -R wip", tags: []string{"@foo", "@bar"}, want: false},
		{basket: "-R foo -R bar|wip", tags: []string{"@foo", "@bar"}, want: true},
		{basket: "-R foo -X bar", tags: []string{"@foo", "@bar"}, want: false},
		{basket: "-X bar -R foo", tags: []string{"@foo", "@bar"}, want: false},
		{basket: "-R foo -X wip", tags: []string{"@foo", "@bar"}, want: true},
		{basket: "-R foo -X foo", tags: []string{"@foo", "@bar"}, want: false},
		{basket: "-R foo|wip -R bar", tags: []string{"@foo", "@bar"}, want: true},

		{basket: "-r foo -X @wip", name: "foo", tags: []string{"@foo"}, want: true},
		{basket: "-r foo -X @foo", name: "foo", tags: []string{"@foo"}, want: false},
		{basket: "-x foo -X @foo", name: "foo", tags: []string{"@foo"}, want: false},
		{basket: "-x foo -X @wip", name: "foo", tags: []string{"@foo"}, want: false},
		{basket: "-x foo -R @foo", name: "foo", tags: []string{"@foo"}, want: false},
	}

	for _, tt := range tests {
		b, err := ntt.NewBasket("testBasket", strings.Fields(tt.basket)...)
		if err != nil {
			t.Fatal(err)
		}
		actual := b.Match(tt.name, findTags(tt.tags...))
		if actual != tt.want {
			t.Errorf("Basket(%q).Match(%q, %q) = %v, want %v", tt.basket, tt.name, tt.tags, actual, tt.want)
		}
	}
}

// makeTags finds all tags in the given string slice.
func findTags(tags ...string) [][]string {
	return doc.FindAllTags(strings.Join(tags, "\n"))
}
