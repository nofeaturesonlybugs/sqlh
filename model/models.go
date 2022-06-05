package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/set/path"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
	"github.com/nofeaturesonlybugs/sqlh/schema"
)

// Models is a registry of Models and methods to manipulate them.
type Models struct {
	//
	// Mapper defines how SQL column names map to fields in Go structs.
	Mapper *set.Mapper
	//
	// Grammar defines the SQL grammar to use for SQL generation.
	Grammar grammar.Grammar
	//
	// Models is a map of Go types to Model instances.  This member is automatically
	// instantiated during calls to Register().
	Models map[reflect.Type]*Model
	//
	// StructTag specifies the struct tag name to use when inspecting types
	// during register.  If not set will default to "model".
	StructTag string
}

// Register adds a Go type to the Models instance.
//
// Register is not goroutine safe; implement locking in the store or application level if required.
//
// When Register is called with a type T the following registrations are made:
//	T, *T, []T, & []*T
//
// As a convenience register can be called with a reflect.Type as the value.
func (me *Models) Register(value interface{}, opts ...interface{}) {
	tagName := me.StructTag
	if tagName == "" {
		tagName = "model"
	}
	//
	if me.Models == nil {
		me.Models = make(map[reflect.Type]*Model)
	}
	//
	var typ reflect.Type
	var typInfo set.TypeInfo
	switch sw := value.(type) {
	case reflect.Type:
		typ = sw
		typInfo = set.TypeCache.StatType(typ)
	default:
		typ = reflect.TypeOf(value)
		typInfo = set.TypeCache.Stat(value) // Consider creating a local type cache to the Models type.
	}
	if _, ok := me.Models[typ]; ok {
		return // Already registered.
	}
	//
	// Get the table name from embedded TableName field.
	var tableName string
	for _, opt := range opts {
		if tn, ok := opt.(TableName); ok {
			tableName = string(tn)
		}
	}
	if tableName == "" {
		for _, field := range typInfo.StructFields {
			if field.Type == typeTableName {
				tableName = field.Tag.Get(tagName)
				break
			}
		}
	}
	if tableName == "" {
		panic("no table name; call Register with a TableName value or embed TableName into your struct")
	}
	//
	// Now map the columns.
	mapping := me.Mapper.Map(value)
	//
	// key is the Columns for the table's primary key.
	// unique is the slice of unique indexes on the table.
	// columns are the non-primary key columns and includes columns in unique.
	key, unique, columns := []schema.Column{}, []schema.Index{}, []schema.Column{}
	//
	// The following slices keep track of column names in the database.
	//	autoKeyNames, keyNames
	//		+ Primary key column names.
	//		+ Composite primary keys (keys using multiple fields) are supported.
	//		+ autoKeyNames are columns automatically populated by the database such as auto incrementing integer keys.
	//	autoInsertNames, autoUpdateNames, autoInsertUpdateNames
	//		+ Columns automatically populated by the database such as created or modified timestamps.
	//		+ autoInsertUpdateNames is UNIQUE( UNION( autoInsertNames, autoUpdateNames ) ).
	//	columnNames
	//		+ All other column names that need to be explicitly set during insert/update operations.
	//
	// NB: auto* columns are not currently limited to any specific type.
	autoKeyNames, autoInsertNames, autoUpdateNames, autoInsertUpdateNames, keyNames, columnNames := []string{}, []string{}, []string{}, []string{}, []string{}, []string{}
	for _, name := range mapping.Keys {
		field := mapping.StructFields[name]
		if field.Type == typeTableName {
			// Leave as empty case to ensure embedded TableName is not used for column information.
		} else {
			// Create the Column type.
			column := schema.Column{
				Name:   name,
				GoType: reflect.Zero(field.Type).Interface(),
				// TODO SqlType
			}
			// Get the struct field tag and then classify the column accordingly.
			tag := field.Tag.Get(tagName)
			if tag == "key" || strings.HasPrefix(tag, "key,") {
				// tag=key or tag=key,auto is a primary key field.
				key = append(key, column)
				if strings.Contains(tag, ",auto") {
					autoKeyNames = append(autoKeyNames, name)
				} else {
					keyNames = append(keyNames, name)
				}
			} else if insert, update := strings.Contains(tag, "inserted"), strings.Contains(tag, "updated"); insert || update {
				// inserted or updated signals the column is populated on insert or update statements respectively.
				if insert {
					autoInsertNames = append(autoInsertNames, name)
				}
				if update {
					autoUpdateNames = append(autoUpdateNames, name)
				}
				if insert || update {
					autoInsertUpdateNames = append(autoInsertUpdateNames, name)
				}
			} else {
				// All other columns are explicitly set during queries.
				columns = append(columns, column)
				columnNames = append(columnNames, name)
			}
			if strings.Contains(tag, "unique") {
				// unique signals the column is part of a unique index.
				// TODO Currently only single column unique indexes are supported; should also support multi-column.
				// TODO The above comment is a lie -- indexes aren't supported at all yet.
				index := schema.Index{
					Name:      "", // TODO Index name.
					Columns:   []schema.Column{column},
					IsPrimary: false,
					IsUnique:  true,
				}
				unique = append(unique, index)
			}
		}
	}
	//
	// Determine the model's save mode.
	var saveMode SaveMode
	switch {
	case len(keyNames) > 0:
		saveMode = Upsert
	case len(autoKeyNames) == 0:
		saveMode = Insert
	default:
		saveMode = InsertOrUpdate
	}
	var insertUpdatePaths []path.ReflectPath
	for _, keyName := range autoKeyNames {
		insertUpdatePaths = append(insertUpdatePaths, mapping.ReflectPaths[keyName])
	}
	//
	// Merge autoKeyNames into autoInsertNames as those keys are generated during insert statements.
	autoInsertNames = append(autoKeyNames, autoInsertNames...)
	// Create table struct.
	table := schema.Table{
		Name: tableName,
		PrimaryKey: schema.Index{
			Name:      "", // TODO Index name.
			Columns:   key,
			IsPrimary: true,
			IsUnique:  true,
		},
		Unique:  unique,
		Columns: columns,
	}
	// Create model struct.
	model := &Model{
		Table:             table,
		Statements:        statements.Table{},
		SaveMode:          saveMode,
		InsertUpdatePaths: insertUpdatePaths,
		Mapping:           mapping,
	}
	// Fill in query statements.
	// NB: Ignore errors here as we'll handle when a query is nil for a model in our other functions.
	model.Statements.Insert, _ = me.Grammar.Insert(tableName, append(keyNames, columnNames...), autoInsertNames)
	model.Statements.Update, _ = me.Grammar.Update(tableName, columnNames, append(autoKeyNames, keyNames...), autoUpdateNames)
	model.Statements.Delete, _ = me.Grammar.Delete(tableName, append(autoKeyNames, keyNames...))
	model.Statements.Upsert, _ = me.Grammar.Upsert(tableName, columnNames, keyNames, autoInsertUpdateNames)
	//
	// We want to be able to look up the model by the original type T passed to this function
	// as well as []T.
	me.Models[typ] = model
	me.Models[reflect.PtrTo(typ)] = model
	me.Models[reflect.SliceOf(typ)] = model
	me.Models[reflect.SliceOf(reflect.PtrTo(typ))] = model
}

// Lookup returns the model associated with the value.
func (me *Models) Lookup(value interface{}) (m *Model, err error) {
	if me == nil {
		err = errors.NilReceiver()
		return
	}
	var ok bool
	t := reflect.TypeOf(value)
	if m, ok = me.Models[t]; ok {
		return
	}
	err = errors.Errorf("%T not registered", value)
	return
}

// Insert attempts to persist values via INSERTs.
func (me *Models) Insert(Q sqlh.IQueries, value interface{}) error {
	var model *Model
	var query *statements.Query
	var binding QueryBinding
	var err error
	if model, err = me.Lookup(value); err != nil {
		return errors.Go(err)
	} else if query = model.Statements.Insert; query == nil {
		return errors.Go(ErrUnsupported).Tag("INSERT", fmt.Sprintf("%T", value))
	}
	//
	binding = model.BindQuery(me.Mapper, query)
	if err = binding.Query(Q, value); err != nil {
		return errors.Go(err)
	}
	//
	return nil
}

// Update attempts to persist values via UPDATESs.
func (me *Models) Update(Q sqlh.IQueries, value interface{}) error {
	var model *Model
	var query *statements.Query
	var binding QueryBinding
	var err error
	if model, err = me.Lookup(value); err != nil {
		return errors.Go(err)
	} else if query = model.Statements.Update; query == nil {
		return errors.Go(ErrUnsupported).Tag("UPDATE", fmt.Sprintf("%T", value))
	}
	//
	binding = model.BindQuery(me.Mapper, query)
	if err = binding.Query(Q, value); err != nil {
		return errors.Go(err)
	}
	//
	return nil
}

// Save inspects the incoming model and delegates to Insert, Update, or Upsert method
// according to the model's SaveMode value, which is determined during registration.
//
// Models with at least one key field defined as "key" (i.e. not "key,auto") use Upsert.
//
// Models with zero "key" and "key,auto" fields use Insert.
//
// Otherwise the model has only "key,auto" fields and will use Update if any such field
// is a non-zero value and Insert otherwise.
//
// If value is a slice []M then the first element is inspected to determine which of
// Insert, Update, or Upsert is applied to the entire slice.
func (me *Models) Save(Q sqlh.IQueries, value interface{}) error {
	model, err := me.Lookup(value)
	if err != nil {
		return errors.Go(err)
	}
	switch model.SaveMode {
	case Insert:
		return me.Insert(Q, value)
	case Upsert:
		return me.Upsert(Q, value)
	case InsertOrUpdate:
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice:
			if v.Len() == 0 {
				return nil
			}
			for v = v.Index(0); v.Kind() == reflect.Ptr; v = v.Elem() {
				if v.IsNil() {
					return errors.Go(ErrUnsupported).Tag("nil pointer", fmt.Sprintf("%v %v", v.Type(), v.Interface()))
				}
			}
		case reflect.Ptr:
			for ; v.Kind() == reflect.Ptr; v = v.Elem() {
				if v.IsNil() {
					return errors.Go(ErrUnsupported).Tag("nil pointer", fmt.Sprintf("%v %v", v.Type(), v.Interface()))
				}
			}
		}
		// TODO Possibly add support for an InsertUpdater interface
		//
		var keyValue reflect.Value
		for _, path := range model.InsertUpdatePaths {
			keyValue = path.Value(v)
			if !keyValue.IsZero() {
				// A non-zero field value means update.
				return me.Update(Q, value)
			}
		}
		return me.Insert(Q, value)
	}
	// Currently it _should_be_ impossible for this to occur.  The first thing this method
	// does is find the associated model and -- if not found -- returns error.  Any model that
	// is found is (currently at least) guaranteed to have a SaveMode corresponding to one
	// of the above values.
	return errors.Go(ErrUnsupported).Tag("SAVE", fmt.Sprintf("%T", value))
}

// Upsert attempts to persist values via UPSERTs.
//
// Upsert only works on primary keys that are defined as "key"; in other words columns tagged with "key,auto"
// are not used in the generated query.
//
// Upsert only supports primary keys; currently there is no support for upsert on UNIQUE indexes that are
// not primary keys.
func (me *Models) Upsert(Q sqlh.IQueries, value interface{}) error {
	var model *Model
	var query *statements.Query
	var binding QueryBinding
	var err error
	if model, err = me.Lookup(value); err != nil {
		return errors.Go(err)
	} else if query = model.Statements.Upsert; query == nil {
		return errors.Go(ErrUnsupported).Tag("UPSERT", fmt.Sprintf("%T", value))
	}
	//
	binding = model.BindQuery(me.Mapper, query)
	if err = binding.Query(Q, value); err != nil {
		return errors.Go(err)
	}
	//
	return nil
}
