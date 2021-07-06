package grammar

import (
	"fmt"
	"strings"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

// Postgres is an instantiated grammar for PostgreSQL.
var Postgres Grammar = &PostgresGrammar{}

// PostgresGrammar defines a grammar for Postgres.
type PostgresGrammar struct {
}

// ParamN returns the string for parameter N where N is zero-based and return value is one-based.
func (me *PostgresGrammar) ParamN(n int) string {
	return fmt.Sprintf("$%v", n+1)
}

// Delete returns the query type for deleting from the table.
func (me *PostgresGrammar) Delete(table string, keys []string) (*statements.Query, error) {
	var keySize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if keySize = len(keys); keySize == 0 {
		return nil, errors.Go(ErrKeysRequired).Tag("table", table).Tag("SQL", "DELETE")
	}
	rv := &statements.Query{
		Arguments: make([]string, keySize),
	}
	wheres := make([]string, keySize)
	for k, key := range keys {
		wheres[k] = key + " = " + me.ParamN(k)
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
func (me *PostgresGrammar) Insert(table string, columns []string, auto []string) (*statements.Query, error) {
	var colSize int
	if table == "" {
		return nil, errors.Go(ErrTableRequired)
	} else if colSize = len(columns); colSize == 0 {
		return nil, errors.Go(ErrColumnsRequired).Tag("table", table).Tag("SQL", "INSERT")
	}
	rv := &statements.Query{
		Arguments: make([]string, colSize),
	}
	values := make([]string, colSize)
	for k, column := range columns {
		values[k] = me.ParamN(k)
		rv.Arguments[k] = column
	}
	//
	parts := []string{
		"INSERT INTO " + table,
		"\t\t( " + strings.Join(columns, ", ") + " )",
		"\tVALUES",
		"\t\t( " + strings.Join(values, ", ") + " )",
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
func (me *PostgresGrammar) Update(table string, columns []string, keys []string, auto []string) (*statements.Query, error) {
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
	//
	sets, wheres := make([]string, colSize), make([]string, keySize)
	for k, column := range columns {
		sets[k] = column + " = " + me.ParamN(k)
		rv.Arguments[k] = column
	}
	for k, key := range keys {
		total := colSize + k
		wheres[k] = key + " = " + me.ParamN(total)
		rv.Arguments[total] = key
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
		rv.Expect = statements.ExpectRowOrNone
	}
	rv.SQL = strings.Join(parts, "\n")
	return rv, nil
}

// Upsert returns the query type for upserting (INSERT|UPDATE) a record in a table.
func (me *PostgresGrammar) Upsert(table string, columns []string, keys []string, auto []string) (*statements.Query, error) {
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
	// INSERT...VALUES portion.
	values := make([]string, sizeInsert)
	for k := range rv.Arguments {
		values[k] = me.ParamN(k)
	}
	// Create an AS alias for the target table.
	alias := "dest"
	// Only columns are used for the DO UPDATE portion of the query.
	updateColumns := make([]string, colSize)
	whereColumns := make([]string, colSize)
	for k, column := range columns {
		updateColumns[k] = column + " = EXCLUDED." + column
		whereColumns[k] = alias + "." + column + " <> EXCLUDED." + column
	}
	//
	parts := []string{
		"INSERT INTO " + table + " AS " + alias,
		"\t\t( " + strings.Join(rv.Arguments, ", ") + " )",
		"\tVALUES",
		"\t\t( " + strings.Join(values, ", ") + " )",
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
