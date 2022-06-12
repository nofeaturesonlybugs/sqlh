package model

import "reflect"

// TableName represents a database table name.  Embed the TableName type into a struct and set
// the appropriate struct tag to configure the table name.
type TableName string

var (
	typeTableName = reflect.TypeOf(TableName(""))
)
