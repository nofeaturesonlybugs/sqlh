package grammar_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

func TestDefaultGrammar(t *testing.T) {
	chk := assert.New(t)
	//
	g := grammar.Sqlite
	//
	{
		// inserts
		columns := []string{"a", "b", "c"}
		auto := []string(nil)
		query, err := g.Insert("foo", columns, auto)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "INSERT INTO foo\n\t\t( a, b, c )\n\tVALUES\n\t\t( ?, ?, ? )"
		chk.Equal(expect, query.SQL)
		chk.Equal([]string{"a", "b", "c"}, query.Arguments)
		chk.Empty(query.Scan)
		// insert with returning
		columns = []string{"a", "b", "c"}
		auto = []string{"x", "y", "z"}
		query, err = g.Insert("foo", columns, auto)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "INSERT INTO foo\n\t\t( a, b, c )\n\tVALUES\n\t\t( ?, ?, ? )\n\tRETURNING x, y, z"
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
		query, err := g.Update("foo", columns, keys, auto)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "UPDATE foo SET\n\t\ta = ?,\n\t\tb = ?,\n\t\tc = ?\n\tWHERE\n\t\tx = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append(append([]string{}, columns...), keys...), query.Arguments)
		chk.Empty(query.Scan)
		// update with returning
		columns = []string{"a", "b", "c"}
		keys = []string{"x"}
		auto = []string{"y", "z"}
		query, err = g.Update("foo", columns, keys, auto)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "UPDATE foo SET\n\t\ta = ?,\n\t\tb = ?,\n\t\tc = ?\n\tWHERE\n\t\tx = ?\n\tRETURNING y, z"
		chk.Equal(expect, query.SQL)
		chk.Equal(append(append([]string{}, columns...), keys...), query.Arguments)
		chk.Equal([]string{"y", "z"}, query.Scan)
	}
	//
	{
		// deletes
		keys := []string{"x"}
		query, err := g.Delete("foo", keys)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect := "DELETE FROM foo\n\tWHERE\n\t\tx = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
		//
		// composite key
		keys = []string{"x", "y", "z"}
		query, err = g.Delete("foo", keys)
		chk.NoError(err)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		expect = "DELETE FROM foo\n\tWHERE\n\t\tx = ? AND y = ? AND z = ?"
		chk.Equal(expect, query.SQL)
		chk.Equal(append([]string{}, keys...), query.Arguments)
		chk.Empty(query.Scan)
	}
}

func TestDefaultReturnsErrors(t *testing.T) {
	chk := assert.New(t)
	//
	var err error
	table, columns, keys := "mytable", []string{"a", "b", "c"}, []string{"x", "y"}
	g := grammar.Sqlite
	// Missing table name.
	_, err = g.Delete("", keys)
	chk.Error(err)
	_, err = g.Insert("", columns, nil)
	chk.Error(err)
	_, err = g.Update("", columns, keys, nil)
	chk.Error(err)
	_, err = g.Upsert("", columns, keys, nil)
	chk.Error(err)
	// Missing keys.
	_, err = g.Delete(table, nil)
	chk.Error(err)
	_, err = g.Update(table, columns, nil, nil)
	chk.Error(err)
	_, err = g.Upsert(table, columns, nil, nil)
	chk.Error(err)
	// Missing columns.
	_, err = g.Insert(table, nil, nil)
	chk.Error(err)
	_, err = g.Update(table, nil, keys, nil)
	chk.Error(err)
	_, err = g.Upsert(table, nil, keys, nil)
	chk.Error(err)
}

func TestDefaultGrammarUpsert(t *testing.T) {
	chk := assert.New(t)
	//
	g := grammar.Sqlite
	{ // single key, no auto
		columns := []string{"a", "b", "c"}
		keys := []string{"key"}
		query, err := g.Upsert("foo", columns, keys, nil)
		chk.NoError(err)
		chk.NotNil(query)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		parts := []string{
			"INSERT INTO foo\n\t\t( key, a, b, c )\n\tVALUES\n\t\t( ?, ?, ?, ? )",
			"\tON CONFLICT( key ) DO UPDATE SET",
			"\t\tfoo.a = EXCLUDED.a, foo.b = EXCLUDED.b, foo.c = EXCLUDED.c",
			"\t\tWHERE (\n\t\t\tfoo.a <> EXCLUDED.a OR foo.b <> EXCLUDED.b OR foo.c <> EXCLUDED.c\n\t\t)",
		}
		expect := strings.Join(parts, "\n")
		chk.Equal(expect, query.SQL)
		args := append([]string{}, keys...)
		args = append(args, columns...)
		chk.Equal(args, query.Arguments)
	}
	{ // composite key, no auto
		columns := []string{"a", "b", "c"}
		keys := []string{"key1", "key2", "key3"}
		query, err := g.Upsert("foo", columns, keys, nil)
		chk.NoError(err)
		chk.NotNil(query)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		parts := []string{
			"INSERT INTO foo\n\t\t( key1, key2, key3, a, b, c )\n\tVALUES\n\t\t( ?, ?, ?, ?, ?, ? )",
			"\tON CONFLICT( key1, key2, key3 ) DO UPDATE SET",
			"\t\tfoo.a = EXCLUDED.a, foo.b = EXCLUDED.b, foo.c = EXCLUDED.c",
			"\t\tWHERE (\n\t\t\tfoo.a <> EXCLUDED.a OR foo.b <> EXCLUDED.b OR foo.c <> EXCLUDED.c\n\t\t)",
		}
		expect := strings.Join(parts, "\n")
		chk.Equal(expect, query.SQL)
		args := append([]string{}, keys...)
		args = append(args, columns...)
		chk.Equal(args, query.Arguments)
	}
	{ // composite key, has auto
		columns := []string{"a", "b", "c"}
		keys := []string{"key1", "key2", "key3"}
		auto := []string{"created", "modified"}
		query, err := g.Upsert("foo", columns, keys, auto)
		chk.NoError(err)
		chk.NotNil(query)
		chk.NotEmpty(query.SQL)
		chk.NotEmpty(query.Arguments)
		parts := []string{
			"INSERT INTO foo\n\t\t( key1, key2, key3, a, b, c )\n\tVALUES\n\t\t( ?, ?, ?, ?, ?, ? )",
			"\tON CONFLICT( key1, key2, key3 ) DO UPDATE SET",
			"\t\tfoo.a = EXCLUDED.a, foo.b = EXCLUDED.b, foo.c = EXCLUDED.c",
			"\t\tWHERE (\n\t\t\tfoo.a <> EXCLUDED.a OR foo.b <> EXCLUDED.b OR foo.c <> EXCLUDED.c\n\t\t)",
			"\tRETURNING created, modified",
		}
		expect := strings.Join(parts, "\n")
		chk.Equal(expect, query.SQL)
		args := append([]string{}, keys...)
		args = append(args, columns...)
		chk.Equal(args, query.Arguments)
	}
	{
		// various errors...
		var query *statements.Query
		var err error
		//
		// empty table
		query, err = g.Upsert("", []string{"a", "b", "c"}, []string{"key"}, nil)
		chk.Error(err)
		chk.Nil(query)
		// empty columns
		query, err = g.Upsert("foo", nil, []string{"key"}, nil)
		chk.Error(err)
		chk.Nil(query)
		// empty keys
		query, err = g.Upsert("foo", []string{"a", "b", "c"}, nil, nil)
		chk.Error(err)
		chk.Nil(query)
	}
}
