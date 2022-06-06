package sqlh_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/errors"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/examples"
)

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

func TestScanner_Select_Errors(t *testing.T) {
	db, mock, _ := sqlmock.New()
	type Dest struct {
		A int
		B int
	}
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{},
	}

	t.Run("nil dest", func(t *testing.T) {
		chk := assert.New(t)
		err := scanner.Select(db, nil, "select * from test")
		chk.Error(err)
	})
	t.Run("invalid dest", func(t *testing.T) {
		chk := assert.New(t)
		var d map[string]interface{}
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
	})
	t.Run("readonly dest", func(t *testing.T) {
		chk := assert.New(t)
		var d Dest
		err := scanner.Select(db, d, "select * from test")
		chk.Error(err)
	})
	t.Run("struct column mismatch", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"X", "Y"})
		dataRows.AddRow(1, 2)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d Dest
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("struct column mismatch", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"X", "Y"})
		dataRows.AddRow(1, 2)
		dataRows.AddRow(3, 4)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d []Dest
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("struct rows first scan fails", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"A", "B"})
		dataRows.AddRow("a", "b")
		dataRows.AddRow(3, 4)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d []Dest
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("struct rows second scan fails", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"A", "B"})
		dataRows.AddRow(3, 4)
		dataRows.AddRow("a", "b")
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d []Dest
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("scalar rows first scan fails", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow("abc")
		dataRows.AddRow(4)
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d []int
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("scalar rows second scan fails", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow(3)
		dataRows.AddRow("abc")
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows).RowsWillBeClosed()

		var d []int
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("scalar rows error", func(t *testing.T) {
		chk := assert.New(t)

		dataRows := sqlmock.NewRows([]string{"n"})
		dataRows.AddRow(3)
		dataRows.AddRow(4)
		dataRows.RowError(0, errors.Errorf("oops"))

		var d []int
		err := scanner.Select(db, &d, "select * from test")
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
}

func TestScanner_ScanRows_Errors(t *testing.T) {
	type Dest struct {
		A int
		B int
	}
	//
	db, mock, _ := sqlmock.New()
	//
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{},
	}

	t.Run("nil dest", func(t *testing.T) {
		chk := assert.New(t)

		mock.ExpectQuery("select +").
			WillReturnRows(sqlmock.NewRows([]string{"a"}).AddRow(10)).
			RowsWillBeClosed()

		rows, err := db.Query("select * from foo")
		chk.NoError(err)
		err = scanner.ScanRows(rows, nil)
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("invalid dest", func(t *testing.T) {
		chk := assert.New(t)

		var n int
		mock.ExpectQuery("select +").
			WillReturnRows(sqlmock.NewRows([]string{"a"}).AddRow(10)).
			RowsWillBeClosed()

		rows, err := db.Query("select * from foo")
		chk.NoError(err)
		err = scanner.ScanRows(rows, &n)
		chk.Error(err)
		chk.NoError(mock.ExpectationsWereMet())
	})
	t.Run("readonly dest", func(t *testing.T) {
		chk := assert.New(t)
		var d Dest
		err := scanner.Select(db, d, "select * from test")
		chk.Error(err)
	})
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

func TestScanner_Select(t *testing.T) {
	type SimpleStruct struct {
		Message string
		Number  int
	}
	//
	type NestedInnerStruct struct {
		Id       int       `json:"id"`
		Created  time.Time `json:"created"`
		Modified time.Time `json:"modified"`
	}
	type NestedOuterStruct struct {
		NestedInnerStruct
		Message string `json:"message"`
		Number  int    `json:"value" db:"num"`
	}
	type PointerOuterStruct struct {
		*NestedInnerStruct
		Message string `json:"message"`
		Number  int    `json:"value" db:"num"`
	}

	//
	// Some examples return times from time generator.
	tg := examples.TimeGenerator{}
	times := []time.Time{tg.Next(), tg.Next(), tg.Next(), tg.Next()}
	//
	type SelectTest struct {
		Name         string
		Example      examples.Example
		Dest         reflect.Type
		DestPointers reflect.Type
		Expect       interface{}
		Scanner      *sqlh.Scanner
	}
	tests := []SelectTest{
		{
			Name:         "slice-struct",
			Example:      examples.ExSimpleMapper,
			Dest:         reflect.TypeOf([]SimpleStruct(nil)),
			DestPointers: reflect.TypeOf([]*SimpleStruct(nil)),
			Expect: []SimpleStruct{
				{Message: "Hello, World!", Number: 42},
				{Message: "So long!", Number: 100},
			},
			Scanner: &sqlh.Scanner{
				Mapper: &set.Mapper{},
			},
		},
		{
			Name:         "slice-struct-with-nesting",
			Example:      examples.ExNestedStruct,
			Dest:         reflect.TypeOf([]NestedOuterStruct(nil)),
			DestPointers: reflect.TypeOf([]*NestedOuterStruct(nil)),
			Expect: []NestedOuterStruct{
				{
					NestedInnerStruct: NestedInnerStruct{Id: 1, Created: times[0], Modified: times[1]},
					Message:           "Hello, World!",
					Number:            42,
				},
				{
					NestedInnerStruct: NestedInnerStruct{Id: 2, Created: times[2], Modified: times[3]},
					Message:           "So long!",
					Number:            100,
				},
			},
			Scanner: &sqlh.Scanner{
				Mapper: &set.Mapper{
					Elevated: set.NewTypeList(NestedInnerStruct{}),
					Tags:     []string{"db", "json"},
				},
			},
		},
		{
			Name:         "slice-struct-with-pointer-nesting",
			Example:      examples.ExNestedStruct,
			Dest:         reflect.TypeOf([]PointerOuterStruct(nil)),
			DestPointers: reflect.TypeOf([]*PointerOuterStruct(nil)),
			Expect: []NestedOuterStruct{
				{
					NestedInnerStruct: NestedInnerStruct{Id: 1, Created: times[0], Modified: times[1]},
					Message:           "Hello, World!",
					Number:            42,
				},
				{
					NestedInnerStruct: NestedInnerStruct{Id: 2, Created: times[2], Modified: times[3]},
					Message:           "So long!",
					Number:            100,
				},
			},
			Scanner: &sqlh.Scanner{
				Mapper: &set.Mapper{
					Elevated: set.NewTypeList(NestedInnerStruct{}),
					Tags:     []string{"db", "json"},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			chk := assert.New(t)
			db, err := examples.Connect(test.Example)
			chk.NoError(err)

			dest := reflect.New(test.Dest).Interface()

			err = test.Scanner.Select(db, dest, "select * from mytable")
			chk.NoError(err)

			// There's different ways we can check for equality here but we'll just see if
			// what we have encodes the same as what we expect.
			expect, err := json.Marshal(test.Expect)
			chk.NoError(err)
			actual, err := json.Marshal(dest)
			chk.NoError(err)
			chk.Equal(expect, actual)
		})
		t.Run(test.Name+"-pointers", func(t *testing.T) {
			t.Parallel()

			chk := assert.New(t)
			db, err := examples.Connect(test.Example)
			chk.NoError(err)

			dest := reflect.New(test.DestPointers).Interface()

			err = test.Scanner.Select(db, dest, "select * from mytable")
			chk.NoError(err)

			// There's different ways we can check for equality here but we'll just see if
			// what we have encodes the same as what we expect.
			expect, err := json.Marshal(test.Expect)
			chk.NoError(err)
			actual, err := json.Marshal(dest)
			chk.NoError(err)
			chk.Equal(expect, actual)
		})
		t.Run(test.Name+"-scan-rows", func(t *testing.T) {
			t.Parallel()

			chk := assert.New(t)
			db, err := examples.Connect(test.Example)
			chk.NoError(err)

			dest := reflect.New(test.Dest).Interface()

			rows, err := db.Query("select * from mytable")
			chk.NoError(err)
			defer rows.Close()

			err = test.Scanner.ScanRows(rows, dest)
			chk.NoError(err)

			// There's different ways we can check for equality here but we'll just see if
			// what we have encodes the same as what we expect.
			expect, err := json.Marshal(test.Expect)
			chk.NoError(err)
			actual, err := json.Marshal(dest)
			chk.NoError(err)
			chk.Equal(expect, actual)
		})

	}
}
