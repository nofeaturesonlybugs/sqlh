package model

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/nofeaturesonlybugs/set"

	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
)

// QueryBinding binds a model together with a specific query.
type QueryBinding struct {
	mapper *set.Mapper
	model  *Model
	query  *statements.Query
}

// Query accepts either a single model M or a slice of models []M.  It then
// runs and returns the result of QueryOne or QuerySlice.
func (me QueryBinding) Query(q sqlh.IQueries, value interface{}) error {
	if reflect.Slice == reflect.TypeOf(value).Kind() {
		if err := me.QuerySlice(q, value); err != nil {
			return err
		}
	} else if err := me.QueryOne(q, value); err != nil {
		return err
	}
	return nil
}

// QueryOne runs the query against a single instance of the model.
func (me QueryBinding) QueryOne(q sqlh.IQueries, value interface{}) error {
	args, scans := make([]interface{}, len(me.query.Arguments)), make([]interface{}, len(me.query.Scan))
	//
	// Create our prepared mapping.  Note that if the calls to Plan() succeed then we do
	// not need to check errors on the following statement for that plan.
	prepared, err := me.mapper.Prepare(value)
	if err != nil {
		return err // TODO sentinal or wrap?
	}
	if err := prepared.Plan(me.query.Arguments...); err != nil {
		return err
	}
	_, _ = prepared.Fields(args)
	if err := prepared.Plan(me.query.Scan...); err != nil {
		return err
	}
	_, _ = prepared.Assignables(scans)
	//
	// If no scans then use Exec().
	if len(me.query.Scan) == 0 {
		if _, err := q.Exec(me.query.SQL, args...); err != nil {
			return err
		}
		return nil
	}
	//
	row := q.QueryRow(me.query.SQL, args...)
	// NB: The error conditions are separated for code coverage purposes.
	if err := row.Scan(scans...); err != nil {
		if err != sql.ErrNoRows {
			return err
		} else if err == sql.ErrNoRows && me.query.Expect != statements.ExpectRowOrNone {
			return err
		}
	}
	return nil
}

// QuerySlice runs the query against a slice of model instances.
func (me QueryBinding) QuerySlice(q sqlh.IQueries, values interface{}) error {
	v := reflect.ValueOf(values)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("values expects a slice; got %T", values) // TODO+NB Sentinal error
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
	// If the calls to Plan succeed then further calls to Fields or Assignables will not error.
	preparedArgs, err := me.mapper.Prepare(v.Index(0))
	if err != nil {
		return err // TODO sentinal? wrap?
	}
	preparedScans := preparedArgs.Copy()
	if err = preparedArgs.Plan(me.query.Arguments...); err != nil {
		return err
	}
	if err = preparedScans.Plan(me.query.Scan...); err != nil {
		return err
	}
	args, scans := make([]interface{}, len(me.query.Arguments)), make([]interface{}, len(me.query.Scan))
	//
	// If original parameter supports transactions...
	if txer, ok := q.(sqlh.IBegins); ok {
		if tx, err = txer.Begin(); err != nil {
			return err
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
			return err
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
			elem := v.Index(k)
			preparedArgs.Rebind(elem)
			_, _ = preparedArgs.Fields(args)
			//
			if _, err = Exec(args...); err != nil {
				return err
			}
		}
	} else {
		for k := 0; k < size; k++ {
			elem := v.Index(k).Interface()
			preparedArgs.Rebind(elem)
			_, _ = preparedArgs.Fields(args)
			preparedScans.Rebind(elem)
			_, _ = preparedScans.Assignables(scans)
			//
			row = QueryRow(args...)
			if err = row.Scan(scans...); err != nil {
				if err != sql.ErrNoRows {
					return err
				} else if err == sql.ErrNoRows && me.query.Expect != statements.ExpectRowOrNone {
					return err
				}
			}
		}
	}

	//
	// If we opened a transaction then attempt to commit.
	if tx != nil {
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
