package model

// SaveMode describes how a model should be saved when passed to Models.Save method.
type SaveMode int

const (
	_ SaveMode = iota // Skip zero value and make it unusable.

	// Models with zero key fields can only be inserted.
	Insert

	// Models with only key,auto fields use insert or update depending
	// on the current values of the key fields.  If any key,auto field
	// is not the zero value then update otherwise insert.
	InsertOrUpdate

	// Models with at least one key field that is not auto must use upsert.
	Upsert
)
