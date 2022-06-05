package model_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/hobbled"
	"github.com/nofeaturesonlybugs/sqlh/model"
	"github.com/nofeaturesonlybugs/sqlh/model/examples"
	"github.com/stretchr/testify/assert"
)

// Test is a test function and descriptive name.
type Test struct {
	Name string
	Test func(t *testing.T)
}

// ModelQueryTest describes each test and allows us to compose our tests.
type ModelQueryTest struct {
	Name        string
	DBWrapper   hobbled.Wrapper
	MockFn      func(mock sqlmock.Sqlmock)
	ModelsFn    func(Q sqlh.IQueries, Data interface{}) error
	Data        interface{}
	ExpectError bool
}

// ModelQueryTestSlice is a slice of Meta objects.
type ModelQueryTestSlice []ModelQueryTest

// Tests returns a []Test from a ModelQueryMetaSlice.
func (me ModelQueryTestSlice) Tests() []Test {
	tests := []Test{}
	for _, queryTest := range me {
		test := Test{
			Name: fmt.Sprintf("%v: %v", queryTest.DBWrapper.String(), queryTest.Name),
			Test: func(qt ModelQueryTest) func(t *testing.T) {
				return func(t *testing.T) {
					chk := assert.New(t)
					dbm, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					chk.NoError(err)
					//
					db := qt.DBWrapper.WrapDB(dbm)
					qt.MockFn(mock)
					//
					err = qt.ModelsFn(db, qt.Data)
					if qt.ExpectError {
						chk.Error(err)
					} else {
						chk.NoError(err)
					}
					chk.NoError(mock.ExpectationsWereMet())
				}
			}(queryTest),
		}
		tests = append(tests, test)
	}
	return tests
}

func TestModels_DoubleRegister(t *testing.T) {
	// Double registering a model should hit an early return in Register()
	//
	mdb := examples.NewModels()
	mdb.Register(&examples.Address{})
}
func TestModels_RegisterPanicNoTableName(t *testing.T) {
	// Models must have a table name when registering.
	chk := assert.New(t)
	//
	type T struct{}
	recovered := false
	recoverFunc := func() {
		if r := recover(); r != nil {
			recovered = true
		}
	}
	mdb := examples.NewModels()
	func() {
		defer recoverFunc()
		mdb.Register(&T{})
	}()
	chk.True(recovered)
}

func TestModels_RegisterTableNameOption(t *testing.T) {
	// We can provide table name either by embedding in struct or passing as argument to Register()
	chk := assert.New(t)
	//
	type A struct {
		model.TableName `model:"table_a"`
		Name            string `db:"name"`
	}
	type B struct {
		Name string `db:"name"`
	}
	//
	mdb := examples.NewModels()
	mdb.Register(&A{})
	mdb.Register(&B{}, model.TableName("table_b"))
	//
	mdl, err := mdb.Lookup(&A{})
	chk.NoError(err)
	chk.NotNil(mdl)
	chk.Equal("table_a", mdl.Table.Name)
	//
	mdl, err = mdb.Lookup(&B{})
	chk.NoError(err)
	chk.NotNil(mdl)
	chk.Equal("table_b", mdl.Table.Name)
}

func TestModels_TableNameTypeChecking(t *testing.T) {
	// model.TableName is a type string; want to make sure we can differentiate it from other strings.
	//
	tn, str := model.TableName(""), "Hello, World!"
	//
	i := interface{}(tn)
	switch i.(type) {
	case string:
		t.Fatalf("tn hits string case")
	case model.TableName:
	}
	//
	i = interface{}(str)
	switch i.(type) {
	case model.TableName:
		t.Fatalf("tn hits mode.TableName case")
	case string:
	}
}

func TestModels_NilReceiver(t *testing.T) {
	// Some functions have nil receiver checks.
	chk := assert.New(t)
	//
	var mdb *model.Models
	m, err := mdb.Lookup(nil)
	chk.Error(err)
	chk.Nil(m)
}

func TestModels_NilArguments(t *testing.T) {
	// Passing nil to some functions hits early return statements.
	chk := assert.New(t)
	//
	db, _, _ := sqlmock.New()
	mdb := examples.NewModels()
	m, err := mdb.Lookup(nil)
	chk.Error(err)
	chk.Nil(m)
	//
	err = mdb.Insert(db, nil)
	chk.Error(err)
	err = mdb.Update(db, nil)
	chk.Error(err)
	err = mdb.Upsert(db, nil)
	chk.Error(err)
}

func TestModelsUnsupported(t *testing.T) {
	chk := assert.New(t)
	//
	type T struct{}
	m := &model.Models{
		Grammar: grammar.Sqlite,
		Mapper:  &set.Mapper{},
	}
	m.Register(&T{}, model.TableName("panic_table_T"))
	//
	var err error
	db, _, _ := sqlmock.New()
	err = m.Insert(db, &T{})
	chk.Error(err)
	chk.Equal(model.ErrUnsupported, errors.Original(err))
	err = m.Update(db, &T{})
	chk.Error(err)
	chk.Equal(model.ErrUnsupported, errors.Original(err))
	err = m.Upsert(db, &T{})
	chk.Error(err)
	chk.Equal(model.ErrUnsupported, errors.Original(err))
}

func TestModelsQueriesError(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NotNil(db)
	chk.NotNil(mock)
	chk.NoError(err)
	//
	mock.ExpectQuery("INSERT+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Insert(db, &examples.Address{})
	chk.Error(err)
	//
	mock.ExpectQuery("INSERT+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Insert(db, []*examples.Address{{}, {}})
	chk.Error(err)
	//
	mock.ExpectQuery("UPDATE+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Update(db, &examples.Address{})
	chk.Error(err)
	//
	mock.ExpectQuery("UPDATE+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Update(db, []*examples.Address{{}, {}})
	chk.Error(err)
	//
	mock.ExpectQuery("UPSERT+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Upsert(db, &examples.Upsertable{})
	chk.Error(err)
	//
	mock.ExpectQuery("UPSERT+").WillReturnError(errors.Errorf("some error"))
	err = examples.Models.Upsert(db, []*examples.Upsertable{{}, {}})
	chk.Error(err)
}

// MakeModelQueryTestsForCompositeKeyNoAuto builds a slice of Test types to test a model with
// composite primary key and no auto-updating fields.
func MakeModelQueryTestsForCompositeKeyNoAuto() []Test {
	// Relationship is a model with a composite primary key and no fields that auto update.
	// Such a model might exist for relationship tables.
	type Relationship struct {
		model.TableName `json:"-" model:"relationship"`
		//
		LeftId  int `json:"left_id" db:"left_fk" model:"key"`
		RightId int `json:"right_id" db:"right_fk" model:"key"`
		// Such a table might have other columns.
		Toggle bool `json:"toggle"`
	}
	//
	models := &model.Models{
		Mapper: &set.Mapper{
			Join: "_",
			Tags: []string{"db", "json"},
		},
		Grammar: grammar.Postgres,
	}
	models.Register(&Relationship{})
	//
	relate := &Relationship{
		LeftId:  -1,
		RightId: -10,
		Toggle:  false,
	}
	relateSlice := []*Relationship{
		{
			LeftId:  1,
			RightId: 10,
			Toggle:  false,
		},
		{
			LeftId:  2,
			RightId: 20,
			Toggle:  true,
		},
		{
			LeftId:  -3,
			RightId: -30,
			Toggle:  false,
		},
	} //
	SQLInsert := strings.Join([]string{
		"INSERT INTO relationship",
		"\t\t( left_fk, right_fk, toggle )",
		"\tVALUES",
		"\t\t( $1, $2, $3 )",
	}, "\n")
	SQLUpdate := strings.Join([]string{
		"UPDATE relationship SET",
		"\t\ttoggle = $1",
		"\tWHERE",
		"\t\tleft_fk = $2 AND right_fk = $3",
	}, "\n")
	SQLUpsert := strings.Join([]string{
		"INSERT INTO relationship AS dest",
		"\t\t( left_fk, right_fk, toggle )",
		"\tVALUES",
		"\t\t( $1, $2, $3 )",
		"\tON CONFLICT( left_fk, right_fk ) DO UPDATE SET",
		"\t\ttoggle = EXCLUDED.toggle",
		"\t\tWHERE (",
		"\t\t\tdest.toggle <> EXCLUDED.toggle",
		"\t\t)",
	}, "\n")

	//
	meta := []ModelQueryTest{
		//
		// INSERTS
		{
			Name:      "insert single with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLInsert).WithArgs(-1, -10, false).WillReturnError(fmt.Errorf("relationship error"))
			},
			ExpectError: true,
			ModelsFn:    models.Insert,
			Data:        relate,
		},
		{
			Name:      "insert slice with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				prepare := mock.ExpectPrepare(SQLInsert)
				prepare.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
				mock.ExpectRollback()
			},
			ExpectError: true,
			ModelsFn:    models.Insert,
			Data:        relateSlice,
		},
		{
			Name:      "insert slice with error",
			DBWrapper: hobbled.NoBegin,
			MockFn: func(mock sqlmock.Sqlmock) {
				prepare := mock.ExpectPrepare(SQLInsert)
				prepare.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Insert,
			Data:        relateSlice,
		},
		{
			Name:      "insert slice with error",
			DBWrapper: hobbled.NoBeginNoPrepare,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLInsert).WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLInsert).WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLInsert).WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Insert,
			Data:        relateSlice,
		},
		//
		// UPDATES
		{
			Name:      "update single with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLUpdate).WithArgs(false, -1, -10).WillReturnError(fmt.Errorf("relationship error"))
			},
			ExpectError: true,
			ModelsFn:    models.Update,
			Data:        relate,
		},
		{
			Name:      "update slice with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				prepare := mock.ExpectPrepare(SQLUpdate)
				prepare.ExpectExec().WithArgs(false, 1, 10).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(true, 2, 20).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(false, -3, -30).WillReturnError(fmt.Errorf("relationship slice error"))
				mock.ExpectRollback()
			},
			ExpectError: true,
			ModelsFn:    models.Update,
			Data:        relateSlice,
		},
		{
			Name:      "update slice with error",
			DBWrapper: hobbled.NoBegin,
			MockFn: func(mock sqlmock.Sqlmock) {
				prepare := mock.ExpectPrepare(SQLUpdate)
				prepare.ExpectExec().WithArgs(false, 1, 10).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(true, 2, 20).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(false, -3, -30).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Update,
			Data:        relateSlice,
		},
		{
			Name:      "update slice with error",
			DBWrapper: hobbled.NoBeginNoPrepare,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLUpdate).WithArgs(false, 1, 10).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLUpdate).WithArgs(true, 2, 20).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLUpdate).WithArgs(false, -3, -30).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Update,
			Data:        relateSlice,
		},
		//
		// UPSERT
		{
			Name:      "upsert single with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLUpsert).WithArgs(-1, -10, false).WillReturnError(fmt.Errorf("relationship error"))
			},
			ExpectError: true,
			ModelsFn:    models.Upsert,
			Data:        relate,
		},
		{
			Name:      "upsert slice with error",
			DBWrapper: hobbled.Passthru,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				prepare := mock.ExpectPrepare(SQLUpsert)
				prepare.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
				mock.ExpectRollback()
			},
			ExpectError: true,
			ModelsFn:    models.Upsert,
			Data:        relateSlice,
		},
		{
			Name:      "upsert slice with error",
			DBWrapper: hobbled.NoBegin,
			MockFn: func(mock sqlmock.Sqlmock) {
				prepare := mock.ExpectPrepare(SQLUpsert)
				prepare.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				prepare.ExpectExec().WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Upsert,
			Data:        relateSlice,
		},
		{
			Name:      "upsert slice with error",
			DBWrapper: hobbled.NoBeginNoPrepare,
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(SQLUpsert).WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLUpsert).WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(SQLUpsert).WithArgs(-3, -30, false).WillReturnError(fmt.Errorf("relationship slice error"))
			},
			ExpectError: true,
			ModelsFn:    models.Upsert,
			Data:        relateSlice,
		},
	}
	return ModelQueryTestSlice(meta).Tests()
}

func TestModels_Suite(t *testing.T) {
	for _, test := range MakeModelQueryTestsForCompositeKeyNoAuto() {
		t.Run(test.Name, test.Test)
	}
}
