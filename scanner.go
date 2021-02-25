package sqlh

import (
	"reflect"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
)

// Scanner facilitates scanning query results into destinations.
type Scanner struct {
	*set.Mapper
}

// Select uses Q to run the query string with args and scans results into dest.
func (me *Scanner) Select(Q IQuery, dest interface{}, query string, args ...interface{}) error {
	if me == nil {
		return errors.NilReceiver()
	} else if Q == nil {
		return errors.NilMember("Q").Type(Q)
	}
	rows, err := Q.Query(query, args...)
	if err != nil {
		return errors.Go(err)
	}
	defer rows.Close()
	if err = me.ScanRows(rows, dest); err != nil {
		return errors.Go(err)
	}
	return nil
}

// ScanRows scans rows from R into dest.
func (me *Scanner) ScanRows(R IRows, dest interface{}) error {
	if R == nil {
		return errors.NilArgument("R").Type(R)
	}
	defer R.Close()
	//
	var D *set.Value
	var columns []string
	var err error
	//
	if me == nil {
		return errors.NilReceiver()
	} else if me.Mapper == nil {
		return errors.NilMember("Mapper").Type(me.Mapper)
	} else if dest == nil {
		return errors.NilArgument("dest").Type(dest)
	} else if columns, err = R.Columns(); err != nil {
		return errors.Go(err)
	} else if D = set.V(dest); !D.CanWrite {
		return errors.Errorf("Dest is not writable; call %T.Rows with address-of dest.", me)
	} else if !D.IsSlice {
		return errors.Errorf("Dest should be slice but is type= %T", dest)
	} else if !D.ElemTypeInfo.IsStruct {
		return errors.Errorf("Dest.Element should be type struct but is type= %v", D.ElemTypeInfo.Type)
	}
	//
	var assignables []interface{}
	// D is a slice; E is then an element instance that can be appended to D.
	e := reflect.New(D.ElemType).Interface()
	E := set.V(e)
	bound := me.Mapper.Bind(e)
	// Want to use our existing bound element; otherwise we're creating and discarding one.
	if R.Next() {
		if assignables, err = bound.Assignables(columns); err != nil {
			// An error here indicates an unknown column is in the result set; i.e. a column for
			// which there is no mapping.
			return errors.Go(err)
		} else if err = R.Scan(assignables...); err != nil {
			return errors.Go(err)
		}
		set.Panics.Append(D, E)
	}
	for R.Next() {
		// Create new element E; ignore error because we already know the call succeeds.
		e = reflect.New(D.ElemType).Interface()
		E.Rebind(e)
		bound.Rebind(e)
		// Get the assignable interface{} values; again we already know the call succeeds.
		assignables, _ = bound.Assignables(columns)
		if err = R.Scan(assignables...); err != nil {
			return errors.Go(err)
		}
		// We use set.Panics.Append() because we have reasonable guarantee the panic can not occur.
		// Our elements E are created by calling D.NewElem() and should always be of correct type.
		set.Panics.Append(D, E)
	}
	if err = R.Err(); err != nil {
		return errors.Go(err)
	}
	return nil
}
