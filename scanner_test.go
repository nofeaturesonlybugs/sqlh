package sqlh_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
)

// rowsColumnsError is an IRows where the Columns() call returns an error.
type rowsColumnsError struct{}

func (r rowsColumnsError) Close() error                   { return nil }
func (r rowsColumnsError) Columns() ([]string, error)     { return nil, errors.Errorf("columns error") }
func (r rowsColumnsError) Err() error                     { return nil }
func (r rowsColumnsError) Next() bool                     { return false }
func (r rowsColumnsError) Scan(dest ...interface{}) error { return nil }

func TestScanner_StructSliceQueryError(t *testing.T) {
	// Tests Query(...) returns non-nil error when dest is []struct.
	//
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(mock)
	chk.NotNil(db)
	{ // When dest is []struct and query returns error
		mock.ExpectQuery("select +").WillReturnError(errors.Errorf("[]struct query error"))
		type Dest struct {
			A string
		}
		scanner := &sqlh.Scanner{
			Mapper: &set.Mapper{},
		}
		var d []Dest
		err = scanner.Select(db, &d, "select * from table")
		chk.Error(err)
	}
}

func TestScanner_StructSliceRowsErrNonNil(t *testing.T) {
	// Tests that *sql.Rows.Err() != nil after the for...*sql.Rows.Next() {} loop when dest is a []struct.
	//
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(mock)
	chk.NotNil(db)
	{
		// When dest is []struct and *sql.Rows.Err() is non-nil
		rows := sqlmock.NewRows([]string{"A"}).
			AddRow("a").AddRow("b").AddRow("c").
			RowError(2, errors.Errorf("[]struct *sql.Rows.Err() is non-nil"))
		mock.ExpectQuery("select +").WillReturnRows(rows)
		type Dest struct {
			A string
		}
		scanner := &sqlh.Scanner{
			Mapper: &set.Mapper{},
		}
		var d []Dest
		err = scanner.Select(db, &d, "select * from table")
		chk.Error(err)
	}
}

func TestScanner_SelectScalarTime(t *testing.T) {
	// Tests selecting into a dest of time.Time.
	//
	chk := assert.New(t)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	{
		// scalar time
		rv := time.Now()
		dataRows := sqlmock.NewRows([]string{"tm"})
		dataRows.AddRow(rv)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		//
		var n time.Time
		err = scanner.Select(db, &n, "select max(*) as tm from foo")
		chk.NoError(err)
		chk.True(rv.Equal(n))
	}
}

func TestScanner_SelectScalarSlice(t *testing.T) {
	// Tests selecting into a dest of []time.Time.
	chk := assert.New(t)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	{
		// scalar time
		row1 := time.Now()
		row2 := row1.Add(-1 * time.Hour)
		dataRows := sqlmock.NewRows([]string{"tm"})
		dataRows.AddRow(row1)
		dataRows.AddRow(row2)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		//
		var n []time.Time
		err = scanner.Select(db, &n, "select max(*) as tm from foo")
		chk.NoError(err)
		chk.True(row1.Equal(n[0]))
		chk.True(row2.Equal(n[1]))
	}
}

func TestScanner_SingleStruct(t *testing.T) {
	// Tests various errors when scanning into a single struct.
	chk := assert.New(t)
	//
	type Dest struct {
		A string `json:"a"`
		B int    `json:"b"`
	}
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	{
		// query returns error
		mock.ExpectQuery("select (.+)").WillReturnError(errors.Errorf("oops"))
		//
		var d Dest
		err = scanner.Select(db, &d, "select a, b from foo")
		chk.Error(err)
	}
	{ // columns error
		dataRows := sqlmock.NewRows([]string{"a", "b", "c"})
		dataRows.AddRow("Hello", 42, "not found")
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		//
		var d Dest
		err = scanner.Select(db, &d, "select a, b, c from foo")
		chk.Error(err)
	}
	{ // scan error
		dataRows := sqlmock.NewRows([]string{"a", "b"})
		dataRows.AddRow("Hello", "asdf")
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		//
		var d Dest
		err = scanner.Select(db, &d, "select a, b, c from foo")
		chk.Error(err)
	}
}

func TestScanner_RowsColumnsError(t *testing.T) {
	chk := assert.New(t)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{},
	}
	rows := rowsColumnsError{}
	var dest []struct{}
	err := scanner.ScanRows(rows, &dest)
	chk.Error(err)
}

func TestScanner_ScanErrors(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	type Dest struct {
		A int
		B int
	}
	var dest []*Dest
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{},
	}
	//
	{ // mismatch column name
		dataRows := sqlmock.NewRows([]string{"X", "Y"})
		dataRows.AddRow(1, 2)
		dataRows.AddRow(3, 4)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		rows, err := db.Query("select * from test")
		chk.NoError(err)
		chk.NotNil(rows)
		defer rows.Close()
		err = scanner.ScanRows(rows, &dest)
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // first scan fails
		dataRows := sqlmock.NewRows([]string{"A", "B"})
		dataRows.AddRow("a", "b")
		dataRows.AddRow(3, 4)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		rows, err := db.Query("select * from test")
		chk.NoError(err)
		chk.NotNil(rows)
		defer rows.Close()
		err = scanner.ScanRows(rows, &dest)
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // second scan fails
		dataRows := sqlmock.NewRows([]string{"A", "B"})
		dataRows.AddRow(3, 4)
		dataRows.AddRow("a", "b")
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		rows, err := db.Query("select * from test")
		chk.NoError(err)
		chk.NotNil(rows)
		defer rows.Close()
		err = scanner.ScanRows(rows, &dest)
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // first scan fails, dest is slice of scalar
		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow("abc")
		dataRows.AddRow(4)
		var n []int
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = scanner.Select(db, &n, "select * from foo")
		chk.Error(err)
	}
	{ // second scan fails, dest is slice of scalar
		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow(3)
		dataRows.AddRow("abc")
		var n []int
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = scanner.Select(db, &n, "select * from foo")
		chk.Error(err)
	}
	{ // second scan fails, dest is slice of scalar
		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow(3)
		dataRows.AddRow("4")
		dataRows.RowError(0, errors.Errorf("oops"))
		var n []int
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = scanner.Select(db, &n, "select * from foo")
		chk.Error(err)
	}
}

func TestScanner_InvalidDest(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	// Select with map dest errors
	dest := map[string]interface{}{}
	err = scanner.Select(db, &dest, "select * from test")
	chk.Error(err)
	// Scan rows with invalid dest.
	var n int
	mock.ExpectQuery("select +").WillReturnRows(sqlmock.NewRows([]string{"a"}).AddRow(10))
	rows, err := db.Query("select * from foo")
	chk.NoError(err)
	err = scanner.ScanRows(rows, &n)
	chk.Error(err)
}
func TestScanner_DestIsNil(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	dataRows := sqlmock.NewRows([]string{"A", "B"})
	dataRows.AddRow(1, 2)
	dataRows.AddRow(3, 4)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	// Select() with nil dest errors
	err = scanner.Select(db, nil, "select * from test")
	chk.Error(err)
	// ScanRows with nil dest errors
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	rows, err := db.Query("select * from test")
	chk.NoError(err)
	defer rows.Close()
	err = scanner.ScanRows(rows, nil)
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}

func TestScanner_DestNotWritable(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	type Dest struct {
		A int
		B int
	}
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	// dest not writable
	var dest []*Dest
	err = scanner.Select(db, dest, "select * from test")
	chk.Error(err)
}

func TestScanner_DestWrongType(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	dataRows := sqlmock.NewRows([]string{"A", "B"})
	dataRows.AddRow(1, 2)
	dataRows.AddRow(3, 4)
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	// dest is not slice
	var di int
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	err = scanner.Select(db, &di, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
	// dest is not slice of struct
	var dslice []int
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	err = scanner.Select(db, &dslice, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}
