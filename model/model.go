package model

import (
	"reflect"

	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
	"github.com/nofeaturesonlybugs/sqlh/schema"
)

// Model relates a Go type to its Table.
type Model struct {
	// Table is the related database table.
	Table schema.Table
	// Statements are the SQL database statements.
	Statements statements.Table

	// V is a *set.Value of a model instance M.
	V set.Value
	// VSlice is a *set.Value of a model slice []M.
	VSlice set.Value

	// Mapping is the column to struct field mapping.
	Mapping set.Mapping

	// BoundMapping is a cached BoundMapping of a zero value for
	// the model.  Gathering query arguments and scan targets will
	// be faster when executing queries by performing the following:
	//	cp := Model.BoundMapping.Copy() // Copy the cached BoundMapping.
	//	cp.Rebind( modelInstance ) 		// Bind the copy to an instance of the model.
	//	cp.Fields( ... )				// Get a slice of query arguments by column name.
	//	cp.Assignable( ... )			// Get a slice of scan targets by column name.
	BoundMapping set.BoundMapping
}

// NewInstance creates an instance of the model's zero value.
func (me *Model) NewInstance() interface{} {
	return reflect.Indirect(reflect.New(me.V.TopValue.Type())).Interface()
}

// NewSlice creates a slice that can hold instances of the model's zero value.
func (me *Model) NewSlice() interface{} {
	return reflect.Indirect(reflect.New(me.VSlice.TypeInfo.Type)).Interface()
}

// BindQuery returns a QueryBinding that facilitates running queries against
// instaces of the model.
func (me *Model) BindQuery(query *statements.Query) QueryBinding {
	return new_query_binding_t(me, query)
}
