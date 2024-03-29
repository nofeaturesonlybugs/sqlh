package model_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model"
	"github.com/nofeaturesonlybugs/sqlh/model/examples"
	"github.com/nofeaturesonlybugs/sqlh/model/statements"
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
	//
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
		bound := modelptr.BindQuery(mdb.Mapper, qu)
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
		bound := modelptr.BindQuery(mdb.Mapper, qu)
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
	models.Register(&Person{})
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
		// Check flow when queryer does not support prepared statements.
		// db that can't prepare statements.
		db := &no_prepare_db{db}
		//
		bound := modelptr.BindQuery(models.Mapper, qu)
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

func TestQueryBinding_QueryOne(t *testing.T) {
	chk := assert.New(t)
	//
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
	models.Register(&Person{})
	modelptr, err := models.Lookup(&Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
		Expect:    statements.ExpectRow,
	}
	{
		// Query expects one row and gets one row.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).AddRow(10))
		person := &Person{}
		err = bound.QueryOne(db, person)
		chk.NoError(err)
		chk.Equal(10, person.Id)
	}
	{
		// Query expects one row and gets no rows.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}))
		person := &Person{}
		err = bound.QueryOne(db, person)
		chk.Error(err)
		chk.Equal(0, person.Id)
	}
	{
		// Query expects one row or none and gets no rows.
		qu.Expect = statements.ExpectRowOrNone
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}))
		person := &Person{Id: 10}
		err = bound.QueryOne(db, person)
		chk.NoError(err)
		chk.Equal(10, person.Id)
	}
}

func TestQueryBinding_QuerySlice_WithPrepare_ExpectRow_GetNone(t *testing.T) {
	chk := assert.New(t)
	//
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
	models.Register(&Person{})
	modelptr, err := models.Lookup([]*Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
		Expect:    statements.ExpectRow,
	}
	{
		// Query expects one row and gets no rows.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectBegin()
		stmt := mock.ExpectPrepare("INSERT+")
		stmt.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		mock.ExpectRollback()
		people := []*Person{{}, {}}
		err = bound.QuerySlice(db, people)
		chk.Error(err)
		chk.Equal(0, people[0].Id)
		chk.Equal(0, people[1].Id)
		chk.NoError(mock.ExpectationsWereMet())
	}
}

func TestQueryBinding_QuerySlice_WithPrepare_ExpectRowOrNone_GetNone(t *testing.T) {
	chk := assert.New(t)
	//
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
	models.Register(&Person{})
	modelptr, err := models.Lookup([]*Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
		Expect:    statements.ExpectRowOrNone,
	}
	{
		// Query expects one row and gets no rows.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectBegin()
		stmt := mock.ExpectPrepare("INSERT+")
		stmt.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		stmt.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		mock.ExpectCommit()
		people := []*Person{{}, {}}
		err = bound.QuerySlice(db, people)
		chk.NoError(err)
		chk.Equal(0, people[0].Id)
		chk.Equal(0, people[1].Id)
		chk.NoError(mock.ExpectationsWereMet())
	}
}

func TestQueryBinding_QuerySlice_NoPrepare_ExpectRow_GetNone(t *testing.T) {
	chk := assert.New(t)
	//
	//
	db, mock, err := sqlmock.New()
	chk.NotNil(db)
	chk.NotNil(mock)
	chk.NoError(err)
	noprepare := &no_prepare_db{db}
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
	models.Register(&Person{})
	modelptr, err := models.Lookup([]*Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
		Expect:    statements.ExpectRow,
	}
	{
		// Query expects one row and gets no rows.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		people := []*Person{{}, {}}
		err = bound.QuerySlice(noprepare, people)
		chk.Error(err)
		chk.Equal(0, people[0].Id)
		chk.Equal(0, people[1].Id)
		chk.NoError(mock.ExpectationsWereMet())
	}
}

func TestQueryBinding_QuerySlice_NoPrepare_ExpectRowOrNone_GetNone(t *testing.T) {
	chk := assert.New(t)
	//
	//
	db, mock, err := sqlmock.New()
	chk.NotNil(db)
	chk.NotNil(mock)
	chk.NoError(err)
	noprepare := &no_prepare_db{db}
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
	models.Register(&Person{})
	modelptr, err := models.Lookup([]*Person{})
	chk.NoError(err)
	//
	qu := &statements.Query{
		SQL:       "INSERT",
		Arguments: []string{"First", "Last"},
		Scan:      []string{"Id"},
		Expect:    statements.ExpectRowOrNone,
	}
	{
		// Query expects one row and gets no rows.
		bound := modelptr.BindQuery(models.Mapper, qu)
		// If query errors.
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		mock.ExpectQuery("INSERT+").WillReturnRows(sqlmock.NewRows([]string{"Id"}).RowError(0, sql.ErrNoRows))
		people := []*Person{{}, {}}
		err = bound.QuerySlice(noprepare, people)
		chk.NoError(err)
		chk.Equal(0, people[0].Id)
		chk.Equal(0, people[1].Id)
		chk.NoError(mock.ExpectationsWereMet())
	}
}

func TestQueryBinding_PreparedMappingErrors(t *testing.T) {
	db, mock, _ := sqlmock.New()
	//
	type Person struct {
		model.TableName `model:"people"`
		Id              int `model:"key,auto"`
		First           string
		Last            string
	}
	var people []Person = []Person{{}, {}}
	var person Person
	//
	models := model.Models{
		Grammar: grammar.Postgres,
		Mapper:  &set.Mapper{},
	}
	models.Register(person)
	modelptr, _ := models.Lookup(person)
	//
	t.Run("not writable", func(t *testing.T) {
		// Model is not writable (can not be prepared by set)
		chk := assert.New(t)
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"First", "Last"},
			Scan:      []string{"Id"},
			Expect:    statements.ExpectRowOrNone,
		}
		bound := modelptr.BindQuery(models.Mapper, qu)
		//
		err := bound.QueryOne(db, person)
		chk.ErrorIs(err, set.ErrReadOnly)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("missing arguments", func(t *testing.T) {
		// Query has arguments not in the struct.
		chk := assert.New(t)
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"Fields", "Not", "Found"},
			Scan:      []string{"Id"},
			Expect:    statements.ExpectRowOrNone,
		}
		bound := modelptr.BindQuery(models.Mapper, qu)
		//
		err := bound.QueryOne(db, &person)
		chk.ErrorIs(err, set.ErrUnknownField)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("missing arguments slice", func(t *testing.T) {
		// Query has arguments not in the struct.
		chk := assert.New(t)
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"Fields", "Not", "Found"},
			Scan:      []string{"Id"},
			Expect:    statements.ExpectRowOrNone,
		}
		bound := modelptr.BindQuery(models.Mapper, qu)
		//
		err := bound.QuerySlice(db, people)
		chk.ErrorIs(err, set.ErrUnknownField)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("missing scan", func(t *testing.T) {
		// Query has scan not in the struct
		chk := assert.New(t)
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"First", "Last"},
			Scan:      []string{"NotFound"},
			Expect:    statements.ExpectRowOrNone,
		}
		bound := modelptr.BindQuery(models.Mapper, qu)
		//
		err := bound.QueryOne(db, &person)
		chk.ErrorIs(err, set.ErrUnknownField)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("missing scan slice", func(t *testing.T) {
		// Query has scan not in the struct
		chk := assert.New(t)
		qu := &statements.Query{
			SQL:       "INSERT",
			Arguments: []string{"First", "Last"},
			Scan:      []string{"NotFound"},
			Expect:    statements.ExpectRowOrNone,
		}
		bound := modelptr.BindQuery(models.Mapper, qu)
		//
		err := bound.QuerySlice(db, people)
		chk.ErrorIs(err, set.ErrUnknownField)
		chk.NoError(mock.ExpectationsWereMet())
	})
}
