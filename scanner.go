package sqlh

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
)

// scannerDestType describes a destination value from the caller.
type scannerDestType int

const (
	destInvalid scannerDestType = iota
	destScalar
	destScalarSlice
	destStruct
	destStructSlice
)

// String returns the Expect value as a string.
func (me scannerDestType) String() string {
	return [...]string{"Invalid", "Scalar", "[]Scalar", "Struct", "[]Struct"}[me]
}

// Scanner facilitates scanning query results into destinations.
type Scanner struct {
	*set.Mapper
}

// inspectValue inspects a query destination and determines if it can be used.
func (me *Scanner) inspectValue(dest interface{}) (V set.Value, T scannerDestType, err error) {
	T = destInvalid
	if dest == nil {
		err = errors.Errorf("dest is nil")
		return
	} else if V = set.V(dest); !V.CanWrite {
		err = errors.Errorf("dest is not writable")
		return
	}
	if V.IsSlice {
		switch dest.(type) {
		case *[]time.Time:
			T = destScalarSlice

		default:
			if V.ElemTypeInfo.IsStruct {
				T = destStructSlice
			} else if V.ElemTypeInfo.IsScalar {
				T = destScalarSlice
			}
		}
	} else {
		switch dest.(type) {
		case *time.Time:
			T = destScalar

		default:
			if V.IsStruct {
				T = destStruct
			} else if V.IsScalar {
				T = destScalar
			}
		}
	}
	if T == destInvalid {
		err = errors.Errorf("unsupported dest %T", dest)
	}
	return
}

// Select uses Q to run the query string with args and scans results into dest.
func (me *Scanner) Select(Q IQueries, dest interface{}, query string, args ...interface{}) error {
	V, T, err := me.inspectValue(dest)
	if err != nil {
		return errors.Go(err)
	}
	switch T {
	case destScalar:
		row := Q.QueryRow(query, args...)
		if err := row.Scan(dest); err != nil {
			return errors.Go(err)
		}

	case destStruct:
		var rows *sql.Rows
		var prepared set.PreparedMapping
		var columns []string
		var err error
		// Why not QueryRow()?  Because *sql.Row does not allow us to get the list of columns which we
		// need for our dynamic Scan().
		if rows, err = Q.Query(query, args...); err != nil {
			return errors.Go(err)
		}
		defer rows.Close()
		if columns, err = rows.Columns(); err != nil {
			return errors.Go(err)
		}
		//
		// Get a prepared mapping and prepare the access plan.  Note that if prepared.Plan()
		// succeeds we know future calls to prepared.Assignables() will succeed and do not
		// need to check those errors.
		if prepared, err = me.Mapper.Prepare(dest); err != nil {
			return errors.Go(err)
		} else if err = prepared.Plan(columns...); err != nil {
			return errors.Go(err)
		}
		assignables := make([]interface{}, len(columns))
		if rows.Next() {
			_, _ = prepared.Assignables(assignables)
			if err = rows.Scan(assignables...); err != nil {
				return errors.Go(err)
			}
		} else {
			// When no rows are returned set dest to the zero value of its type.  Since dest should be a pointer
			// we need to Indirect(ValueOf(dest)) and set TypeOf(dest).Elem().
			reflect.Indirect(reflect.ValueOf(dest)).Set(reflect.Zero(reflect.TypeOf(dest).Elem()))
		}
		if err = rows.Err(); err != nil {
			return errors.Go(err)
		}

	case destScalarSlice:
		fallthrough
	case destStructSlice:
		rows, err := Q.Query(query, args...)
		if err != nil {
			return errors.Go(err)
		}
		defer rows.Close()
		if err = me.scanRows(rows, dest, V, T); err != nil {
			return errors.Go(err)
		}

	}

	return nil
}

// scanRows scans rows is the internal scanRows that assumes dest is safe.
func (me *Scanner) scanRows(R IIterates, dest interface{}, V set.Value, T scannerDestType) error {
	if R != nil {
		defer R.Close()
	}
	var prepared set.PreparedMapping
	var columns []string
	var err error
	//
	switch T {
	case destScalarSlice:
		e := reflect.New(V.ElemType).Interface()
		E := set.V(e)
		if R.Next() {
			if err = R.Scan(e); err != nil {
				return errors.Go(err)
			}
			// While this *can* panic it *should never* panic.  Second famous last words.
			set.Panics.Append(V, E)
		}
		for R.Next() {
			// Create new element E; ignore error because we already know the call succeeds.
			e = reflect.New(V.ElemType).Interface()
			E.Rebind(e)
			if err = R.Scan(e); err != nil {
				return errors.Go(err)
			}
			// While this *can* panic it *should never* panic.  Second famous last words.
			set.Panics.Append(V, E)
		}
		if err = R.Err(); err != nil {
			return errors.Go(err)
		}

	case destStructSlice:
		if columns, err = R.Columns(); err != nil {
			return errors.Go(err)
		}
		//
		assignables := make([]interface{}, len(columns))
		// V is a slice; E is then an element instance that can be appended to V.
		e := reflect.New(V.ElemType)
		// Create new empty slice to reflect.Append(slice, elem) to.  Note on the way
		// out of the function we must assign this slice to V.WriteValue.
		slice := reflect.New(V.Type).Elem()
		//
		// Get a prepared mapping and prepare the access plan.  Note that if prepared.Plan()
		// succeeds we know future calls to prepared.Assignables() will succeed and do not
		// need to check those errors.
		if prepared, err = me.Mapper.Prepare(e); err != nil {
			return errors.Go(err)
		} else if err = prepared.Plan(columns...); err != nil {
			return errors.Go(err)
		}
		//
		// Want to use our existing bound element; otherwise we're creating and discarding one.
		if R.Next() {
			_, _ = prepared.Assignables(assignables)
			if err = R.Scan(assignables...); err != nil {
				return errors.Go(err)
			}
			slice = reflect.Append(slice, e.Elem())
		}
		for R.Next() {
			// Create new element E; ignore error because we already know the call succeeds.
			e = reflect.New(V.ElemType)
			prepared.Rebind(e)
			// Get the assignable values; again we already know the call succeeds.
			_, _ = prepared.Assignables(assignables)
			if err = R.Scan(assignables...); err != nil {
				return errors.Go(err)
			}
			slice = reflect.Append(slice, e.Elem())
		}
		if err = R.Err(); err != nil {
			return errors.Go(err)
		}
		V.WriteValue.Set(slice) // Don't forget to assign the slice we made to caller's dest.
	}

	return nil
}

// ScanRows scans rows from R into dest.
func (me *Scanner) ScanRows(R IIterates, dest interface{}) error {
	if R != nil {
		defer R.Close()
	}
	V, T, err := me.inspectValue(dest)
	if err != nil {
		return errors.Go(err)
	} else if T != destScalarSlice && T != destStructSlice {
		return errors.Errorf("%T.ScanRows expects dest to be address of slice; got %T", me, dest)
	}
	return me.scanRows(R, dest, V, T)
}
