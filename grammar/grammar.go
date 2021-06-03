package grammar

import (
	"fmt"
	"strings"

	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

var (
	// Default is a basic default grammar.
	Default = &Grammar{
		Parameter:   "?",
		ParameterFn: nil,
		Returning:   "RETURNING",
		// TODO Driver
		// Driver: Driver{
		// 	LastInsertID: true,
		// },
	}

	// Postgres is a PostgreSQL grammar.
	Postgres = &Grammar{
		Parameter: "",
		ParameterFn: func(column string, n int) string {
			return fmt.Sprintf("$%v", n+1)
		},
		Returning: "RETURNING",
		// TODO Driver
		// Driver: Driver{
		// 	LastInsertID: true,
		// },
	}
)

// ParameterFunc accepts a column name and the parameter's overall position in the query.
// The return value should be the SQL parameter value to insert into the query.
type ParameterFunc func(string, int) string

// TODO Implement Driver.
// // Driver describes features of the underlying database/sql driver.
// type Driver struct {
// 	// Set to true if the driver supports LastInsertID.  Currently drivers that do not support LastInsertID
// 	// are expected to use a RETURNING clause in queries.
// 	LastInsertID bool
// }

// Grammar creates SQL queries according to the grammar configuration.
type Grammar struct {
	// Parameter and ParameterFn work together to define how parameters are inserted into SQL
	// statements.  If Parameter is non-empty - such as "?" - then its value is used in place
	// of query parameters.  Otherwise ParameterFn is called with the column name and parameter
	// number.
	Parameter   string
	ParameterFn ParameterFunc
	// Returning specifies if the database uses a RETURNING clause to return columns during
	// INSERT/UPDATE queries.  Setting this to an empty string disables this functionality.
	Returning string

	// Driver describes driver features and is important during the execution of queries.
	// Driver Driver // TODO
}

// Delete returns the SQL statement for deleting from the table.  It also returns a slice of column
// names in the order they are used as parameters.
func (me Grammar) Delete(table string, keys []string) (Query *statements.Query) {
	Query = &statements.Query{}
	if table == "" {
		return
	} else if len(keys) == 0 {
		return
	}
	//
	wheres := []string{}
	if me.Parameter != "" {
		for _, key := range keys {
			wheres = append(wheres, key+" = "+me.Parameter)
			Query.Arguments = append(Query.Arguments, key)
		}
	} else if me.ParameterFn != nil {
		for k, key := range keys {
			wheres = append(wheres, key+" = "+me.ParameterFn(key, k))
			Query.Arguments = append(Query.Arguments, key)
		}
	} else {
		panic("Grammar is missing either Parameter or ParameterFn")
	}
	//
	parts := []string{
		"DELETE FROM " + table,
		"\tWHERE",
		"\t\t" + strings.Join(wheres, " AND "),
	}
	Query.SQL = strings.Join(parts, "\n")
	return
}

// Insert returns the SQL statement for inserting into table.  It also returns a slice of column
// names in the order they are used as parameters.
// A RETURNING clause is appended to the statement if auto is non-empty and the grammar supports RETURNING.
func (me Grammar) Insert(table string, columns []string, auto []string) (Query *statements.Query) {
	Query = &statements.Query{}
	if table == "" {
		return
	} else if len(columns) == 0 {
		return
	}
	//
	var values string
	if me.Parameter != "" {
		values = me.Parameter + strings.Repeat(", "+me.Parameter, len(columns)-1)
	} else if me.ParameterFn != nil {
		parts := []string{}
		for k, column := range columns {
			parts = append(parts, me.ParameterFn(column, k))
		}
		values = strings.Join(parts, ", ")
	} else {
		panic("Grammar is missing either Parameter or ParameterFn")
	}
	//
	parts := []string{
		"INSERT INTO " + table,
		"\t\t( " + strings.Join(columns, ", ") + " )",
		"\tVALUES",
		"\t\t( " + values + " )",
	}
	if me.Returning != "" && len(auto) > 0 {
		parts = append(parts, "\t"+me.Returning+" "+strings.Join(auto, ", "))
		Query.Expect = statements.ExpectRow
	}
	Query.SQL, Query.Arguments, Query.Scan = strings.Join(parts, "\n"), append([]string{}, columns...), append([]string{}, auto...)
	return
}

// Update returns the SQL statement for updating a record in a table.  It also returns a slice of column
// names in the order they are used as parameters.
// A RETURNING clause is appended to the statement if auto is non-empty and the grammar supports RETURNING.
func (me Grammar) Update(table string, columns []string, keys []string, auto []string) (Query *statements.Query) {
	Query = &statements.Query{}
	if table == "" {
		return
	} else if len(columns) == 0 {
		return
	} else if len(keys) == 0 {
		return
	}
	//
	sets, wheres := []string{}, []string{}
	if me.Parameter != "" {
		for _, column := range columns {
			sets = append(sets, column+" = "+me.Parameter)
			Query.Arguments = append(Query.Arguments, column)
		}
		for _, key := range keys {
			wheres = append(wheres, key+" = "+me.Parameter)
			Query.Arguments = append(Query.Arguments, key)
		}
	} else if me.ParameterFn != nil {
		n := -1
		for _, column := range columns {
			n++
			sets = append(sets, column+" = "+me.ParameterFn(column, n))
			Query.Arguments = append(Query.Arguments, column)
		}
		for _, key := range keys {
			n++
			wheres = append(wheres, key+" = "+me.ParameterFn(key, n))
			Query.Arguments = append(Query.Arguments, key)
		}
	} else {
		panic("Grammar is missing either Parameter or ParameterFn")
	}
	//
	parts := []string{
		"UPDATE " + table + " SET",
		"\t\t" + strings.Join(sets, ",\n\t\t"),
		"\tWHERE",
		"\t\t" + strings.Join(wheres, " AND "),
	}
	if me.Returning != "" && len(auto) > 0 {
		parts = append(parts, "\t"+me.Returning+" "+strings.Join(auto, ", "))
		Query.Expect = statements.ExpectRow
	}
	Query.SQL, Query.Scan = strings.Join(parts, "\n"), append([]string{}, auto...)
	return
}
