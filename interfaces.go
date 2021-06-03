package sqlh

import (
	"database/sql"
)

// IQueries defines the methods common to types that can run queries.
type IQueries interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// IPrepares defines the methods required to run prepared statements.
type IPrepares interface {
	Prepare(query string) (*sql.Stmt, error)
}

// IIterates defines the methods required for iterating a query result set.
type IIterates interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

// IBegins defines the method(s) required to open a transaction.
type IBegins interface {
	Begin() (*sql.Tx, error)
}
