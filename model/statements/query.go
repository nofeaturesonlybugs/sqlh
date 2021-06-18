package statements

import "fmt"

// Expect is an enum describing what to expect when running a query.
type Expect int

const (
	ExpectNone Expect = iota
	ExpectRow
	ExpectRowOrNone
	ExpectRows
)

// String returns the Expect value as a string.
func (me Expect) String() string {
	return [...]string{"None", "One Row", "One Row or None", "Multiple Rows"}[me]
}

// Query describes a SQL query.
type Query struct {
	// SQL is the query statement.
	SQL string
	// If the query requires arguments then Arguments are the column name arguments in
	// the order expected by the SQL statement.
	Arguments []string
	// If the query returns columns then Scan are the column names in the order
	// to be scanned.
	Scan []string
	// Expect is a hint that indicates if the query returns no rows, one row, or many rows.
	Expect Expect
}

// String describes the Query as a string.
func (me *Query) String() string {
	if me == nil {
		return "Nil Query."
	} else if me.SQL == "" {
		return "Empty Query."
	}
	//
	rv := me.SQL
	if len(me.Arguments) > 0 {
		rv = rv + fmt.Sprintf("\n\tArguments: %v", me.Arguments)
	}
	if len(me.Scan) > 0 {
		rv = rv + fmt.Sprintf("\n\tScan: %v", me.Scan)
	}
	rv = rv + "\n\tExpect: " + me.Expect.String()
	//
	return rv
}
