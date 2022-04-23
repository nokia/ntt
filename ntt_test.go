package ntt_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/nokia/ntt"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/stretchr/testify/assert"
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
		{basket: "-r foo|bar", name: "bar", want: true},
		{basket: "-r foo", name: "bar", want: false},

		{basket: "-x fo.", name: "foo", want: false},
		{basket: "-r foo -r bar", name: "foo", want: false},
		{basket: "-r foo -r bar", name: "foobar", want: true},
		{basket: "-r foo -x bar", name: "foobar", want: false},
		{basket: "-x foo -r bar", name: "foobar", want: false},
		{basket: "-x foo|bar", name: "bar", want: false},
		{basket: "-x foo", name: "bar", want: true},

		{basket: "-X @foo|@bar", want: true},
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
		actual := b.Match(tt.name, doc.FindAllTags(strings.Join(tt.tags, "\n")))
		if actual != tt.want {
			t.Errorf("Basket(%q).Match(%q, %q) = %v, want %v", tt.basket, tt.name, tt.tags, actual, tt.want)
		}
	}
}

func TestSubBaskets(t *testing.T) {
	tests := []struct {
		basket string
		name   string
		tags   []string
		want   bool
	}{
		{basket: "-r foo:-r bar", name: "foo", want: true},
		{basket: "-r foo:-r bar", name: "bar", want: true},
		{basket: "-r foo:-x foo", name: "foo", want: true},
		{basket: "-r foo -x bar:-x foo", name: "foobar", want: false},
		{basket: "-r foo:-x foo:-x foo", name: "foobar", want: true},
		{basket: "-r foo:-x foo:-x bar", name: "foobar", want: true},
	}

	for _, tt := range tests {

		b, err := ntt.NewBasket("testBasket")
		if err != nil {
			t.Fatal(err)
		}

		for _, s := range strings.Split(tt.basket, ":") {
			sb, err := ntt.NewBasket("subBasket", strings.Fields(s)...)
			if err != nil {
				t.Fatal(err)
			}
			b.Baskets = append(b.Baskets, sb)
		}

		actual := b.Match(tt.name, doc.FindAllTags(strings.Join(tt.tags, "\n")))
		if actual != tt.want {
			t.Errorf("Basket(%q).Match(%q, %q) = %v, want %v", tt.basket, tt.name, tt.tags, actual, tt.want)
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("TEST_BASKET", "stable:wip")
	os.Setenv("TEST_BASKET_stable", "-X @wip|@flaky")
	defer func() {
		os.Unsetenv("TEST_BASKET")
		os.Unsetenv("TEST_BASKET_stable")
	}()

	b, err := ntt.NewBasket("testBasket", "-r", "foo")
	if err != nil {
		t.Fatal(err)
	}

	b.LoadFromEnv("TEST_BASKET")
	test := []struct {
		name string
		tags []string
		want bool
	}{
		{name: "foo", want: true},
		{name: "foo", tags: []string{"@wip"}, want: true},
		{name: "foo", tags: []string{"@flaky"}, want: false},
		{name: "bar", want: false},
		{name: "bar", tags: []string{"@wip"}, want: false},
		{name: "bar", tags: []string{"@false"}, want: false},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			actual := b.Match(tt.name, doc.FindAllTags(strings.Join(tt.tags, "\n")))
			if actual != tt.want {
				t.Errorf("Basket.Match(%q, %q) = %v, want %v", tt.name, tt.tags, actual, tt.want)
			}
		})
	}
}

func TestModulePars(t *testing.T) {
	t.Skip()

	tests := []struct {
		file string
		name string
		want []string
	}{
		{name: "", want: nil},
		{name: "xxx", want: nil},
		{name: "xxx.xxx", want: nil},
		{file: "xxx", want: nil},
		{file: "good.parameters", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "A", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "test.A", want: []string{"X=X from good", "Y=Y from good"}},
		{file: "good.parameters", name: "test.B", want: []string{"X=X from good", "Y=Y from B", "Z=Z from B"}},
		{file: "good.parameters", name: "test.C", want: nil},
		{file: "bad.parameters", want: nil},
		{file: "bad.parameters", name: "test.B", want: nil},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			os.Setenv("NTT_PARAMETERS_FILE", tt.file)
			defer os.Unsetenv("NTT_PARAMETERS_FILE")
			got, _ := testModulePars(t, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func testModulePars(t *testing.T, name string) ([]string, error) {
	_, err := ntt.NewSuite("testdata/parameters")
	if err != nil {
		t.Fatalf("NewSuite() failed: %v", err)
	}

	var (
		m map[string]string
	)

	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(s)
	return s, err
}
