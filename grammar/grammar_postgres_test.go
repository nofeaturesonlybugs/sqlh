package grammar_test

import (
	"testing"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/stretchr/testify/assert"
)

func TestPostgresGrammar(t *testing.T) {
	chk := assert.New(t)
	//
	g := grammar.Postgres
	//
	{
		// inserts
		columns := []string{"a", "b", "c"}
		auto := []string{}
		query := g.Insert("foo", columns, auto)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "INSERT INTO foo\n\t\t( a, b, c )\n\tVALUES\n\t\t( $1, $2, $3 )"
		chk.Equal(expect, query.SQL)
		chk.Equal([]string{"a", "b", "c"}, query.Arguments)
		chk.Empty(query.Scan)
		// insert with returning
		columns = []string{"a", "b", "c"}
		auto = []string{"x", "y", "z"}
		query = g.Insert("foo", columns, auto)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "INSERT INTO foo\n\t\t( a, b, c )\n\tVALUES\n\t\t( $1, $2, $3 )\n\tRETURNING x, y, z"
		chk.Equal(expect, query.SQL)
		chk.Equal([]string{"a", "b", "c"}, query.Arguments)
		chk.Equal([]string{"x", "y", "z"}, query.Scan)
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
		expect := "UPDATE foo SET\n\t\ta = $1,\n\t\tb = $2,\n\t\tc = $3\n\tWHERE\n\t\tx = $4"
		chk.Equal(expect, query.SQL)
		chk.Equal(append(append([]string{}, columns...), keys...), query.Arguments)
		chk.Empty(query.Scan)
		// update with returning
		columns = []string{"a", "b", "c"}
		keys = []string{"x"}
		auto = []string{"y", "z"}
		query = g.Update("foo", columns, keys, auto)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "UPDATE foo SET\n\t\ta = $1,\n\t\tb = $2,\n\t\tc = $3\n\tWHERE\n\t\tx = $4\n\tRETURNING y, z"
		chk.Equal(expect, query.SQL)
		chk.Equal(append(append([]string{}, columns...), keys...), query.Arguments)
		chk.Equal([]string{"y", "z"}, query.Scan)
	}
	//
	{
		// deletes
		keys := []string{"x"}
		query := g.Delete("foo", keys)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "DELETE FROM foo\n\tWHERE\n\t\tx = $1"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
		//
		// composite key
		keys = []string{"x", "y", "z"}
		query = g.Delete("foo", keys)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "DELETE FROM foo\n\tWHERE\n\t\tx = $1 AND y = $2 AND z = $3"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
	}
}
