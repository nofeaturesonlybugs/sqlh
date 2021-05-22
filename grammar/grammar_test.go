package grammar_test

import (
	"testing"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
	"github.com/stretchr/testify/assert"
)

func TestGrammarCodeCoverate(t *testing.T) {
	chk := assert.New(t)
	//
	table, columns, keys := "a", []string{"a", "b", "c"}, []string{"z"}
	{
		// Test for early returns when table is empty.
		g := grammar.Grammar{
			Parameter: "?",
		}
		tests := []func() *statements.Query{
			func() *statements.Query {
				return g.Delete("", keys)
			},
			func() *statements.Query {
				return g.Insert("", columns, keys)
			},
			func() *statements.Query {
				return g.Update("", columns, keys, nil)
			},
		}
		for _, test := range tests {
			q := test()
			chk.Equal("", q.SQL)
			chk.Empty(q.Arguments)
			chk.Empty(q.Scan)
		}
	}
	{
		// Test for early returns when columns is empty.
		g := grammar.Grammar{
			Parameter: "?",
		}
		tests := []func() *statements.Query{
			// func() *statements.Query {
			// 	return g.Delete(table, keys)
			// },
			func() *statements.Query {
				return g.Insert(table, []string{}, keys)
			},
			func() *statements.Query {
				return g.Update(table, []string{}, keys, nil)
			},
		}
		for _, test := range tests {
			q := test()
			chk.Equal("", q.SQL)
			chk.Empty(q.Arguments)
			chk.Empty(q.Scan)
		}
	}
	{
		// Test for early returns when keys is empty.
		g := grammar.Grammar{
			Parameter: "?",
		}
		tests := []func() *statements.Query{
			func() *statements.Query {
				return g.Delete(table, []string{})
			},
			// func() *statements.Query {
			// 	return g.Insert(table, columns, []string{})
			// },
			func() *statements.Query {
				return g.Update(table, columns, []string{}, nil)
			},
		}
		for _, test := range tests {
			q := test()
			chk.Equal("", q.SQL)
			chk.Empty(q.Arguments)
			chk.Empty(q.Scan)
		}
	}
	{
		// Test for expected panics.
		g := grammar.Grammar{
			Parameter:   "",
			ParameterFn: nil,
		}
		recovered := false
		recoverFunc := func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}
		tests := []func(){
			func() {
				defer recoverFunc()
				g.Delete(table, keys)
			},
			func() {
				defer recoverFunc()
				g.Insert(table, columns, keys)
			},
			func() {
				defer recoverFunc()
				g.Update(table, columns, keys, nil)
			},
		}
		for _, test := range tests {
			recovered = false
			test()
			chk.Equal(true, recovered)
		}
	}
}
