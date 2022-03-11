package main

import (
	"context"
	"strings"
	"testing"

	"github.com/nokia/ntt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateIDs(t *testing.T) {
	tests := []struct {
		name   string
		ids    string
		policy string
		basket string
		want   []string
	}{
		{name: "new", ids: "", want: []string{"test.A", "test.B", "test2.C"}},
		{name: "new", ids: "test.B", want: []string{"test.B"}},
		{name: "new", ids: "test", want: []string{"test"}},
		{name: "new", ids: "B", want: []string{"B"}},

		{name: "old", policy: "old", ids: "", want: []string{"test.control"}},
		{name: "old", policy: "old", ids: "test.B", want: []string{"test.B"}},
		{name: "old", policy: "old", ids: "test", want: []string{"test"}},
		{name: "old", policy: "old", ids: "B", want: []string{"B"}},

		{name: "basket", basket: "-r A", policy: "old", ids: "", want: nil},
		{name: "basket", basket: "-r A|test", policy: "old", ids: "", want: []string{"test.control"}},
		{name: "basket", basket: "-r A", policy: "old", ids: "test.control", want: []string{"test.control"}},
		{name: "basket", basket: "-r A", ids: "", want: []string{"test.A"}},
		{name: "basket", basket: "-r A", ids: "test.B", want: []string{"test.B"}},
		{name: "basket", basket: "-r A", ids: "test", want: []string{"test"}},
		{name: "basket", basket: "-r A", ids: "B", want: []string{"B"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite, err := ntt.NewSuite("../../testdata/vanilla")
			if err != nil {
				t.Fatal(err)
			}

			var b ntt.Basket
			if tt.basket != "" {
				b, err = ntt.NewBasket("testBasket", strings.Split(tt.basket, " ")...)
				if err != nil {
					t.Fatal(err)
				}
			}

			c := GenerateIDs(context.Background(), strings.Fields(tt.ids), suite.Sources, tt.policy, b)
			var actual []string
			for id := range c {
				actual = append(actual, id)
			}
			assert.Equal(t, tt.want, actual)
		})
	}

}
