package schema

import (
	"fmt"
	"strings"
)

// Index describes a database index.
type Index struct {
	// Name specifies the index name.
	Name string
	// Columns contains the columns in the index.
	Columns []Column
	// IsPrimary is true if the index represents a primary key; if IsPrimary is true then
	// IsUnique is also true.
	IsPrimary bool
	// IsUnique is true if the index represents a unique key.
	IsUnique bool
}

// String describes the index as a string.
func (me Index) String() string {
	if me.Name == "" && len(me.Columns) == 0 && me.IsPrimary == false && me.IsUnique == false {
		return ""
	}
	//
	describe := "unique"
	if me.IsPrimary {
		if len(me.Columns) > 1 {
			describe = "primary composite key"
		} else {
			describe = "primary key"
		}
	}
	name := me.Name
	if name == "" {
		name = "-"
	}
	fields, gotypes, sqltypes := []string{}, []string{}, []string{}
	for _, column := range me.Columns {
		fields = append(fields, column.Name)
		gotypes = append(gotypes, fmt.Sprintf("%T", column.GoType))
		sqltypes = append(sqltypes, column.SqlType)
	}
	return fmt.Sprintf("%v name=%v (%v) go(%v) sql(%v)", describe, name, strings.Join(fields, ","), strings.Join(gotypes, ","), strings.Join(sqltypes, ","))
}
