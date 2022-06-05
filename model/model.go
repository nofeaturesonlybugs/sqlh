package model

import (
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

	// Mapping is the column to struct field mapping.
	Mapping set.Mapping
}

// BindQuery returns a QueryBinding that facilitates running queries against
// instaces of the model.
func (me *Model) BindQuery(mapper *set.Mapper, query *statements.Query) QueryBinding {
	return QueryBinding{
		mapper: mapper,
		model:  me,
		query:  query,
	}
}
