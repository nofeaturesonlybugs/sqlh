package schema

import "fmt"

// Column describes a database column.
type Column struct {
	// Name specifies the column name.
	Name string
	// GoType specifies the type in Go and should be set to a specific Go type.
	GoType interface{}
	// SqlType specifies the type in SQL and can be set to a string describing the SQL type.
	SqlType string
}

// String describes the column as a string.
func (me Column) String() string {
	sqlType := me.SqlType
	if sqlType == "" {
		sqlType = "-"
	}
	return fmt.Sprintf("%v go(%T) sql(%v)", me.Name, me.GoType, sqlType)
}
