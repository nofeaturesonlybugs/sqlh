package model

import (
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/set/path"

	"github.com/nofeaturesonlybugs/sqlh/model/statements"
	"github.com/nofeaturesonlybugs/sqlh/schema"
)

// Model relates a Go type to its Table.
type Model struct {
	// Table is the related database table.
	Table schema.Table

	// Statements are the SQL database statements.
	Statements statements.Table

	// SaveMode is set during model registration and inspected during Models.Save
	// to determine which of Insert, Update, or Upsert operations to use.
	//
	// SaveMode=InsertOrUpdate means InsertUpdatePaths is a non-empty slice of
	// key field traversal information.  The key fields are examined and if
	// any are non-zero values then Update will be used otherwise Insert.
	SaveMode          SaveMode
	InsertUpdatePaths []path.ReflectPath

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
