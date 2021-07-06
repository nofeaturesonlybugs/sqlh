package grammar

import (
	"strings"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

// Sqlite is an instantiated grammar for SQLite.
var Sqlite Grammar = &SqliteGrammar{}

// SqliteGrammar defines a grammar for SQLite v2.35+.
type SqliteGrammar struct {
}

// Delete returns the query type for deleting from the table.
func (me *SqliteGrammar) Delete(table string, keys []string) (*statements.Query, error) {
	var keySize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if keySize = len(keys); keySize == 0 {
		return nil, errors.Go(ErrKeysRequired).Tag("table", table).Tag("SQL", "DELETE")
	}
	rv := &statements.Query{
		Arguments: make([]string, keySize),
	}
	//
	wheres := make([]string, keySize)
	for k, key := range keys {
		wheres[k] = key + " = ?"
		rv.Arguments[k] = key
	}
	//
	parts := []string{
		"DELETE FROM " + table,
		"\tWHERE",
		"\t\t" + strings.Join(wheres, " AND "),
	}
	rv.SQL = strings.Join(parts, "\n")
	return rv, nil
}

// Insert returns the query type for inserting into table.
func (me *SqliteGrammar) Insert(table string, columns []string, auto []string) (*statements.Query, error) {
	var colSize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if colSize = len(columns); colSize == 0 {
		return nil, errors.Go(ErrColumnsRequired).Tag("table", table).Tag("SQL", "INSERT")
	}
	rv := &statements.Query{
		Arguments: make([]string, colSize),
	}
	copy(rv.Arguments[0:], columns)
	values := "?" + strings.Repeat(", ?", colSize-1)
	//
	parts := []string{
		"INSERT INTO " + table,
		"\t\t( " + strings.Join(columns, ", ") + " )",
		"\tVALUES",
		"\t\t( " + values + " )",
	}
	if len(auto) > 0 {
		parts = append(parts, "\tRETURNING "+strings.Join(auto, ", "))
		rv.Scan = append([]string{}, auto...)
		rv.Expect = statements.ExpectRow
	}
	rv.SQL = strings.Join(parts, "\n")
	return rv, nil
}

// Update returns the query type for updating a record in a table.
func (me *SqliteGrammar) Update(table string, columns []string, keys []string, auto []string) (*statements.Query, error) {
	var colSize, keySize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if colSize = len(columns); colSize == 0 {
		return nil, errors.Go(ErrColumnsRequired).Tag("table", table).Tag("SQL", "UPDATE")
	} else if keySize = len(keys); keySize == 0 {
		return nil, errors.Go(ErrKeysRequired).Tag("table", table).Tag("SQL", "UPDATE")
	}
	rv := &statements.Query{
		Arguments: make([]string, colSize+keySize),
	}
	sets, wheres := make([]string, colSize), make([]string, keySize)
	for k, column := range columns {
		sets[k] = column + " = ?"
		rv.Arguments[k] = column
	}
	for k, key := range keys {
		wheres[k] = key + " = ?"
		rv.Arguments[colSize+k] = key
	}
	//
	parts := []string{
		"UPDATE " + table + " SET",
		"\t\t" + strings.Join(sets, ",\n\t\t"),
		"\tWHERE",
		"\t\t" + strings.Join(wheres, " AND "),
	}
	if len(auto) > 0 {
		parts = append(parts, "\tRETURNING "+strings.Join(auto, ", "))
		rv.Scan = append([]string{}, auto...)
		rv.Expect = statements.ExpectRow
	}
	rv.SQL = strings.Join(parts, "\n")
	return rv, nil
}

// Upsert returns the query type for upserting (INSERT|UPDATE) a record in a table.
func (me *SqliteGrammar) Upsert(table string, columns []string, keys []string, auto []string) (*statements.Query, error) {
	var colSize, keySize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if colSize = len(columns); colSize == 0 {
		return nil, errors.Go(ErrColumnsRequired).Tag("table", table).Tag("SQL", "UPDATE")
	} else if keySize = len(keys); keySize == 0 {
		return nil, errors.Go(ErrKeysRequired).Tag("table", table).Tag("SQL", "UPDATE")
	}
	// Both keys + columns are combined for the INSERT portion of the query.
	sizeInsert := colSize + keySize
	rv := &statements.Query{
		Arguments: make([]string, sizeInsert),
	}
	copy(rv.Arguments[0:], keys)
	copy(rv.Arguments[keySize:], columns)
	// Only columns are used for the DO UPDATE portion of the query.
	updateColumns := make([]string, colSize)
	whereColumns := make([]string, colSize)
	for k, column := range columns {
		updateColumns[k] = table + "." + column + " = EXCLUDED." + column
		whereColumns[k] = table + "." + column + " <> EXCLUDED." + column
	}
	//
	// The INSERT...VALUES portion of the query
	values := "?" + strings.Repeat(", ?", sizeInsert-1)
	//
	parts := []string{
		"INSERT INTO " + table,
		"\t\t( " + strings.Join(rv.Arguments, ", ") + " )",
		"\tVALUES",
		"\t\t( " + values + " )",
		"\tON CONFLICT( " + strings.Join(keys, ", ") + " ) DO UPDATE SET",
		"\t\t" + strings.Join(updateColumns, ", "),
		"\t\tWHERE (",
		"\t\t\t" + strings.Join(whereColumns, " OR "),
		"\t\t)",
	}
	if len(auto) > 0 {
		parts = append(parts, "\tRETURNING "+strings.Join(auto, ", "))
		rv.Scan = append([]string{}, auto...)
		rv.Expect = statements.ExpectRowOrNone
	}
	rv.SQL = strings.Join(parts, "\n")
	return rv, nil
}
