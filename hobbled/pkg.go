// Package hobbled allows creation of hobbled or deficient database types to facilitate testing within sqlh.
package hobbled

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nofeaturesonlybugs/sqlh"
)

// Wrapper is a type that wraps around a DB in order to hobble it.
type Wrapper int

const (
	// Passthru does not modify database types given to it.
	Passthru Wrapper = iota
	// NoBegin removes Begin() from database types given to it.
	NoBegin
	// NoBeginNoPrepare removes Begin() and Prepare() from types given to it.
	NoBeginNoPrepare
)

// String describes the wrapper type.
func (me Wrapper) String() string {
	return [...]string{"DB", "DB w/o begin", "DB w/o begin+prepare"}[me]
}

// WrapDB wraps the given DB and returns a sqlh.IQueries instance.
func (me Wrapper) WrapDB(db *sql.DB) sqlh.IQueries {
	switch me {
	case Passthru:
		return db
	case NoBegin:
		return NewWithoutBegin(db)
	case NoBeginNoPrepare:
		return NewWithoutBeginWithoutPrepare(db)
	}
	panic(fmt.Sprintf("unknown %T %v", me, int(me)))
}

// WithoutBegin has no Begin call.
type WithoutBegin interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// NewWithoutBegin returns a WithoutBegin type.
func NewWithoutBegin(db interface{}) WithoutBegin {
	type T struct {
		*canQuery
		*canPrepare
	}
	switch tt := db.(type) {
	case *sql.DB:
		return &T{
			canQuery:   &canQuery{db: tt},
			canPrepare: &canPrepare{db: tt},
		}
	case *sql.Tx:
		return &T{
			canQuery:   &canQuery{tx: tt},
			canPrepare: &canPrepare{tx: tt},
		}
	}
	panic("db is not a *sql.DB or *sql.Tx")
}

// WithoutPrepare has no Prepare calls.
type WithoutPrepare interface {
	Begin() (*sql.Tx, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// NewWithoutPrepare returns a WithoutPrepare type.
func NewWithoutPrepare(db interface{}) WithoutPrepare {
	type T struct {
		*canBegin
		*canQuery
	}
	switch tt := db.(type) {
	case *sql.DB:
		return &T{
			canBegin: &canBegin{db: tt},
			canQuery: &canQuery{db: tt},
		}
	case *sql.Tx:
		panic("*sql.Tx can not be a NoPrepare because NoPrepare has a Begin() method.")
	}
	panic("db is not a *sql.DB or *sql.Tx")
}

// WithoutBeginWithoutPrepare has no Begin or Prepare calls.
type WithoutBeginWithoutPrepare interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// NewWithoutBeginWithoutPrepare returns a WithoutBeginWithoutPrepare type.
func NewWithoutBeginWithoutPrepare(db interface{}) WithoutBeginWithoutPrepare {
	switch tt := db.(type) {
	case *sql.DB:
		return &canQuery{db: tt}
	case *sql.Tx:
		return &canQuery{tx: tt}
	}
	panic("db is not a *sql.DB or *sql.Tx")
}

// canBegin allows the begin function.
type canBegin struct {
	db *sql.DB
}

func (me *canBegin) Begin() (*sql.Tx, error) {
	return me.db.Begin()
}

// canQuery allows the query functions.
type canQuery struct {
	db *sql.DB
	tx *sql.Tx
}

func (me *canQuery) Exec(query string, args ...interface{}) (sql.Result, error) {
	if me.tx != nil {
		return me.tx.Exec(query, args...)
	}
	return me.db.Exec(query, args...)
}
func (me *canQuery) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if me.tx != nil {
		return me.tx.ExecContext(ctx, query, args...)
	}
	return me.db.ExecContext(ctx, query, args...)
}
func (me *canQuery) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if me.tx != nil {
		return me.tx.Query(query, args...)
	}
	return me.db.Query(query, args...)
}
func (me *canQuery) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if me.tx != nil {
		return me.tx.QueryContext(ctx, query, args...)
	}
	return me.db.QueryContext(ctx, query, args...)
}
func (me *canQuery) QueryRow(query string, args ...interface{}) *sql.Row {
	if me.tx != nil {
		return me.tx.QueryRow(query, args...)
	}
	return me.db.QueryRow(query, args...)
}
func (me *canQuery) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if me.tx != nil {
		return me.tx.QueryRowContext(ctx, query, args...)
	}
	return me.db.QueryRowContext(ctx, query, args...)
}

// canPrepare allows the prepare functions.
type canPrepare struct {
	db *sql.DB
	tx *sql.Tx
}

func (me *canPrepare) Prepare(query string) (*sql.Stmt, error) {
	if me.tx != nil {
		return me.tx.Prepare(query)
	}
	return me.db.Prepare(query)
}

func (me *canPrepare) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if me.tx != nil {
		return me.tx.PrepareContext(ctx, query)
	}
	return me.db.PrepareContext(ctx, query)
}
