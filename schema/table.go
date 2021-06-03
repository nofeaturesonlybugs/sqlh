package schema

// Table describes a database table.
type Table struct {
	// Name specifies the database table name.
	Name string
	// Columns represents the table columns.
	Columns []Column
	// PrimaryKey is the index describing the table's primary key if it has one.
	// A primary key - by definition - is a unique index; however it is not also
	// stored in the Unique field.
	PrimaryKey Index
	// Unique is a slice of unique indexes on the table.
	Unique []Index
}

// String describes the table as a string.
func (me Table) String() string {
	rv := ""
	//
	name := me.Name
	if name == "" {
		name = "- (table, name unknown)"
	}
	rv = name
	//
	// primary key
	primary := me.PrimaryKey.String()
	if primary != "" {
		rv = rv + "\n\t" + primary
	}
	//
	// columns
	if len(me.Columns) > 0 {
		rv = rv + "\n\tcolumns"
		for _, column := range me.Columns {
			rv = rv + "\n\t\t" + column.String()
		}
	}
	//
	// unique indexes
	if len(me.Unique) > 0 {
		rv = rv + "\n\tunique indexes"
		for _, index := range me.Unique {
			rv = rv + "\n\t\t" + index.String()
		}
	}
	//
	return rv
}
