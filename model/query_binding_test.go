package model_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model"
	"github.com/nofeaturesonlybugs/sqlh/model/examples"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
	"github.com/stretchr/testify/assert"
)

// no_prepare_db is an IQuery interface that does not support IPrepare.
type no_prepare_db struct {
	db *sql.DB
}

func (me *no_prepare_db) Exec(query string, args ...interface{}) (sql.Result, error) {
	return me.db.Exec(query, args...)
}
func (me *no_prepare_db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return me.db.Query(query, args...)
}
func (me *no_prepare_db) QueryRow(query string, args ...interface{}) *sql.Row {
	return me.db.QueryRow(query, args...)
}

func TestQueryBinding(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NotNil(db)
	chk.NotNil(mock)
	chk.NoError(err)
	//
	mdb := examples.NewModels()
	mdb.Register(examples.Person{})
	//
	model, err := mdb.Lookup(examples.Person{})
	chk.NoError(err)
	chk.NotNil(model)
	modelptr, err := mdb.Lookup(&examples.Person{})
	chk.NoError(err)
	chk.NotNil(modelptr)
	//
	{
		// Check early return conditions for slices.
		// Test qu.Arguments causes the error.
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"first", "last"},
			Scan:      []string{"pk"},
		}
		bound := modelptr.BindQuery(qu)
		chk.NotNil(bound)
		// Early return when not a slice.
		err = bound.QuerySlice(db, int(0))
		chk.Error(err)
		// Early return when nil slice.
		err = bound.QuerySlice(db, []*examples.Person(nil))
		chk.NoError(err)
		// Early return when empty slice.
		err = bound.QuerySlice(db, []*examples.Person{})
		chk.NoError(err)
		// Early return when single element.
		mock.ExpectQuery("INSERT+").WithArgs("", "").WillReturnRows(sqlmock.NewRows([]string{"pk"}).AddRow(10))
		err = bound.QuerySlice(db, []*examples.Person{{}})
		chk.NoError(err)
	}
	{
		// Check the flow path of Query, QueryOne, and QuerySlice
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"first", "last"},
			Scan:      []string{"pk"},
		}
		bound := modelptr.BindQuery(qu)
		chk.NotNil(bound)
		// If begin fails
		mock.ExpectBegin().WillReturnError(errors.Errorf("begin fail"))
		err = bound.QuerySlice(db, []*examples.Person{{}, {}})
		chk.Error(err)
		// If prepare errors.
		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT+").WillReturnError(errors.Errorf("prepare failed"))
		mock.ExpectRollback()
		err = bound.QuerySlice(db, []*examples.Person{{}, {}})
		chk.Error(err)
		// If query errors.
		mock.ExpectBegin()
		prepare := mock.ExpectPrepare("INSERT+")
		prepare.ExpectQuery().WillReturnError(errors.Errorf("query failed"))
		mock.ExpectRollback()
		err = bound.QuerySlice(db, []*examples.Person{{}, {}})
		chk.Error(err)
		// If commit errors.
		mock.ExpectBegin()
		prepare = mock.ExpectPrepare("INSERT+")
		prepare.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"pk"}).AddRow(10))
		prepare.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"pk"}).AddRow(20))
		mock.ExpectCommit().WillReturnError(errors.Errorf("commit failed"))
		err = bound.QuerySlice(db, []*examples.Person{{}, {}})
		chk.Error(err)
	}
	{
		// Check errors with Query, QueryOne, and QuerySlice when non-pointer.
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"first", "last"},
			Scan:      []string{"pk"},
		}
		bound := model.BindQuery(qu)
		chk.NotNil(bound)
		//
		err = bound.QueryOne(db, examples.Person{})
		chk.Error(err)
		//
		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT+")
		mock.ExpectRollback()
		err = bound.QuerySlice(db, []examples.Person{{}, {}})
		chk.Error(err)
	}
}

func TestQueryBinding_NoPrepares(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NotNil(db)
	chk.NotNil(mock)
	chk.NoError(err)
	//
	type Person struct {
		model.TableName `model:"people"`
		Id              int `model:"key,auto"`
		First           string
		Last            string
	}
	models := model.Models{
		Grammar: grammar.Postgres,
		Mapper:  &set.Mapper{},
	}
	models.Register(Person{})
	models.Register(&Person{})
	//
	model, err := models.Lookup(Person{})
	chk.NoError(err)
	//
	modelptr, err := models.Lookup(&Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
	}
	{
		// Check flow when queryer does not support prepared statements and the bind fails.
		// db that can't prepare statements.
		db := &no_prepare_db{db}
		//
		bound := model.BindQuery(qu)
		chk.NotNil(bound)
		// If query errors.
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow(10))
		mock.ExpectQuery("INSERT+").WillReturnError(errors.Errorf("query failed"))
		mock.ExpectRollback()
		err = bound.QuerySlice(db, []Person{{}, {}})
		chk.Error(err)
	}
	{
		// Check flow when queryer does not support prepared statements.
		// db that can't prepare statements.
		db := &no_prepare_db{db}
		//
		bound := modelptr.BindQuery(qu)
		chk.NotNil(bound)
		// If query errors.
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow(10))
		mock.ExpectQuery("INSERT+").WillReturnError(errors.Errorf("query failed"))
		mock.ExpectRollback()
		err = bound.QuerySlice(db, []*Person{{}, {}})
		chk.Error(err)
		// If commit errors.
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow(10))
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow(20))
		mock.ExpectCommit().WillReturnError(errors.Errorf("commit failed"))
		err = bound.QuerySlice(db, []*Person{{}, {}})
		chk.Error(err)
	}
}
