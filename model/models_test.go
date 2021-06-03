package model_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model"
	"github.com/nofeaturesonlybugs/sqlh/model/examples"
	"github.com/stretchr/testify/assert"
)

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
}

func TestModelsPanics(t *testing.T) {
	chk := assert.New(t)
	//
	type T struct{}
	m := &model.Models{
		Grammar: grammar.Default,
		Mapper:  &set.Mapper{},
	}
	m.Register(&T{}, model.TableName("panic_table_T"))
	db, _, _ := sqlmock.New()
	//
	recovered := false
	recoverFunc := func() {
		if r := recover(); r != nil {
			recovered = true
		}
	}
	tests := []func(){
		func() {
			defer recoverFunc()
			m.Insert(db, &T{})
		},
		func() {
			defer recoverFunc()
			m.Update(db, &T{})
		},
	}
	for _, test := range tests {
		recovered = false
		test()
		chk.Equal(true, recovered)
	}
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
}

func TestModels_RegisterSetsVAndVSlice(t *testing.T) {
	chk := assert.New(t)
	//
	mdb := examples.NewModels()
	model, err := mdb.Lookup(&examples.Address{})
	chk.NoError(err)
	chk.NotNil(model)

	v, vs := model.NewInstance(), model.NewSlice()
	if _, ok := v.(*examples.Address); !ok {
		chk.Fail("Model.NewInstance() failed.")
	}
	if _, ok := vs.([]*examples.Address); !ok {
		chk.Fail("Model.NewSlice() failed.")
	}
}
