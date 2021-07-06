package statements

import "strings"

// Table is the collection of Query types to perform CRUD against a table.
type Table struct {
	Delete *Query
	Insert *Query
	Update *Query
	Upsert *Query
}

// String returns the table statements as a friendly string.
func (me Table) String() string {
	parts := []string{
		"INSERT: " + me.Insert.String(),
		"UPDATE: " + me.Update.String(),
		"UPSERT: " + me.Upsert.String(),
		"DELETE: " + me.Delete.String(),
	}
	return strings.Join(parts, "\n")
}
