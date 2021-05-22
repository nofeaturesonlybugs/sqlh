package grammar_test

import (
	"testing"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/stretchr/testify/assert"
)

func TestDefaultGrammar(t *testing.T) {
	chk := assert.New(t)
	//
	g := grammar.Default
	//
	{
		// inserts
		columns := []string{"a", "b", "c"}
		auto := []string(nil)
		query := g.Insert("foo", columns, auto)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "INSERT INTO foo\n\t\t( a, b, c )\n\tVALUES\n\t\t( ?, ?, ? )"
		chk.Equal(expect, query.SQL)
		chk.Equal([]string{"a", "b", "c"}, query.Arguments)
		chk.Empty(query.Scan)
	}
	//
	{
		// updates
		columns := []string{"a", "b", "c"}
		keys := []string{"x"}
		auto := []string(nil)
		query := g.Update("foo", columns, keys, auto)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "UPDATE foo SET\n\t\ta = ?,\n\t\tb = ?,\n\t\tc = ?\n\tWHERE\n\t\tx = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append(append([]string{}, columns...), keys...), query.Arguments)
		chk.Empty(query.Scan)
	}
	//
	{
		// deletes
		keys := []string{"x"}
		query := g.Delete("foo", keys)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "DELETE FROM foo\n\tWHERE\n\t\tx = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
		//
		// composite key
		keys = []string{"x", "y", "z"}
		query = g.Delete("foo", keys)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "DELETE FROM foo\n\tWHERE\n\t\tx = ? AND y = ? AND z = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
	}
}
