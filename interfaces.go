package sqlh

import (
	"database/sql"
)

// IQuery defines the methods common to types that can run queries.
type IQuery interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// IRows defines the methods required for iterating a query result set.
type IRows interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}
