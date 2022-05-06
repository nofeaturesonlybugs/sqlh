package model

import (
	"database/sql"
	"reflect"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

// QueryBinding binds a model with a query to facilitate running the query
// against instances of types described by the model.
type QueryBinding interface {
	// Query accepts either a single model M or a slice of models []M.  It then
	// runs and returns the result of QueryOne or QuerySlice.
	Query(sqlh.IQueries, interface{}) error
	// QueryOne runs the query against a single instance of the model.
	QueryOne(sqlh.IQueries, interface{}) error
	// QuerySlice runs the query against a slice of model instances.
	QuerySlice(sqlh.IQueries, interface{}) error
}

// query_binding_t is the internal implementation of QueryBinding.
type query_binding_t struct {
	model *Model
	query *statements.Query
}

// new_query_binding_t creates a new query binding implementation.
func new_query_binding_t(model *Model, query *statements.Query) QueryBinding {
	rv := &query_binding_t{
		model: model,
		query: query,
	}
	return rv
}

// Query accepts either a single model M or a slice of models []M.  It then
// runs and returns the result of QueryOne or QuerySlice.
func (me *query_binding_t) Query(q sqlh.IQueries, value interface{}) error {
	if reflect.Slice == reflect.TypeOf(value).Kind() {
		if err := me.QuerySlice(q, value); err != nil {
			return errors.Go(err)
		}
	} else if err := me.QueryOne(q, value); err != nil {
		return errors.Go(err)
	}
	return nil
}

// QueryOne runs the query against a single instance of the model.
func (me *query_binding_t) QueryOne(q sqlh.IQueries, value interface{}) error {
	bound, args, scans := me.model.BoundMapping.Copy(), make([]interface{}, len(me.query.Arguments)), make([]interface{}, len(me.query.Scan))
	bound.Rebind(value)
	bound.Fields(me.query.Arguments, args)
	bound.Assignables(me.query.Scan, scans)
	//
	// If no scans then use Exec().
	if len(me.query.Scan) == 0 {
		if _, err := q.Exec(me.query.SQL, args...); err != nil {
			return errors.Go(err)
		}
		return nil
	}
	//
	row := q.QueryRow(me.query.SQL, args...)
	// NB: The error conditions are separated for code coverage purposes.
	if err := row.Scan(scans...); err != nil {
		if err != sql.ErrNoRows {
			return errors.Go(err)
		} else if err == sql.ErrNoRows && me.query.Expect != statements.ExpectRowOrNone {
			return errors.Go(err)
		}
	}
	return nil
}

// QuerySlice runs the query against a slice of model instances.
func (me *query_binding_t) QuerySlice(q sqlh.IQueries, values interface{}) error {
	v := reflect.ValueOf(values)
	if v.Kind() != reflect.Slice {
		return errors.Errorf("values expects a slice; got %T", values) // TODO+NB Sentinal error
	}
	// Size of slice will be helpful here.
	size := v.Len()
	if size == 0 {
		return nil
	} else if size == 1 {
		return me.Query(q, v.Index(0).Interface())
	}
	//
	var tx *sql.Tx
	var stmt *sql.Stmt
	var row *sql.Row
	var err error
	//
	bound, args, scans := me.model.BoundMapping.Copy(), make([]interface{}, len(me.query.Arguments)), make([]interface{}, len(me.query.Scan))
	//
	// If original parameter supports transactions...
	if txer, ok := q.(sqlh.IBegins); ok {
		if tx, err = txer.Begin(); err != nil {
			return errors.Go(err)
		}
		defer tx.Rollback()
		q = tx
	}
	//
	// QueryRowFunc normalizes the query row call so the same logic can be used with or without prepared statements.
	type ExecFunc func(args ...interface{}) (sql.Result, error)
	type QueryRowFunc func(args ...interface{}) *sql.Row
	var QueryRow QueryRowFunc
	var Exec ExecFunc
	//
	// Use prepared statement if possible.
	if pper, ok := q.(sqlh.IPrepares); ok {
		if stmt, err = pper.Prepare(me.query.SQL); err != nil {
			return errors.Go(err)
		}
		defer stmt.Close()
		Exec = stmt.Exec
		QueryRow = stmt.QueryRow
	} else {
		Exec = func(args ...interface{}) (sql.Result, error) {
			return q.Exec(me.query.SQL, args...)
		}
		QueryRow = func(args ...interface{}) *sql.Row {
			return q.QueryRow(me.query.SQL, args...)
		}
	}
	//
	// There's a little bit of copy+paste between both conditions.  Tread carefully when editing the similar portions.
	if len(me.query.Scan) == 0 {
		for k := 0; k < size; k++ {
			bound.Rebind(v.Index(k).Interface())
			bound.Fields(me.query.Arguments, args)
			bound.Assignables(me.query.Scan, scans)
			//
			if _, err = Exec(args...); err != nil {
				return errors.Go(err)
			}
		}
	} else {
		for k := 0; k < size; k++ {
			bound.Rebind(v.Index(k).Interface())
			bound.Fields(me.query.Arguments, args)
			bound.Assignables(me.query.Scan, scans)
			//
			row = QueryRow(args...)
			if err = row.Scan(scans...); err != nil {
				if err != sql.ErrNoRows {
					return errors.Go(err)
				} else if err == sql.ErrNoRows && me.query.Expect != statements.ExpectRowOrNone {
					return errors.Go(err)
				}
			}
		}
	}

	//
	// If we opened a transaction then attempt to commit.
	if tx != nil {
		if err = tx.Commit(); err != nil {
			return errors.Go(err)
		}
	}
	return nil
}
