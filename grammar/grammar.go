package grammar

import (
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

// Grammar creates SQL queries for a specific database engine.
type Grammar interface {
	// Delete returns the query type for deleting from the table.
	Delete(table string, keys []string) (*statements.Query, error)
	// Insert returns the query type for inserting into table.
	Insert(table string, columns []string, auto []string) (*statements.Query, error)
	// Update returns the query type for updating a record in a table.
	Update(table string, columns []string, keys []string, auto []string) (*statements.Query, error)
	// Upsert returns the query type for upserting (INSERT|UPDATE) a record in a table.
	Upsert(table string, columns []string, keys []string, auto []string) (*statements.Query, error)
}

// TODO Implement Driver.
// // Driver describes features of the underlying database/sql driver.
// type Driver struct {
// 	// Set to true if the driver supports LastInsertID.  Currently drivers that do not support LastInsertID
// 	// are expected to use a RETURNING clause in queries.
// 	LastInsertID bool
// }
