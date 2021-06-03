package examples

import (
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
)

// ReturnArgs creates enough return args for the specified number of models in n.
// Columns can be: pk (int), created (time.Time), modified (time.Time)
func ReturnArgs(n int, columns ...string) []driver.Value {
	var rv []driver.Value
	var created time.Time
	for k := 0; k < n; k++ {
		for _, column := range columns {
			switch column {
			case "pk":
				rv = append(rv, rand.Int())
			case "created":
				created = time.Now().Add(-1 * time.Duration((rand.Int() % 3600)) * time.Hour)
				rv = append(rv, created)
			case "modified":
				rv = append(rv, created.Add(time.Duration((rand.Int()%3600))*time.Hour))
			}
		}
	}
	return rv
}

// DB_Insert sets up our mock database for inserting records.
func DB_Insert(models interface{}) (*sql.DB, [][]driver.Value, error) {
	var returning [][]driver.Value
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}
	switch m := models.(type) {
	case *Address:
		returning = append(returning, ReturnArgs(1, "pk", "created", "modified"))
		// Set up expected query.
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		qu := mock.ExpectQuery(strings.Join(parts, "\n"))
		// Set up expected args and return columns.
		args := []driver.Value{m.Street, m.City, m.State, m.Zip}
		rows := sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"})
		rows.AddRow(returning[0]...)
		qu.WithArgs(args...)
		qu.WillReturnRows(rows)
		qu.RowsWillBeClosed()

	case []*Address:
		returning = append(returning, ReturnArgs(1, "pk", "created", "modified"))
		returning = append(returning, ReturnArgs(1, "pk", "created", "modified"))
		// Set up expected query.
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		rows1 := sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"})
		rows1.AddRow(returning[0][0], returning[0][1], returning[0][2])
		rows2 := sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"})
		rows2.AddRow(returning[1][0], returning[1][1], returning[1][2])
		//
		mock.ExpectBegin()
		prepare := mock.ExpectPrepare(strings.Join(parts, "\n"))
		args1 := []driver.Value{m[0].Street, m[0].City, m[0].State, m[0].Zip}
		prepare.ExpectQuery().WithArgs(args1...).WillReturnRows(rows1)
		args2 := []driver.Value{m[1].Street, m[1].City, m[1].State, m[1].Zip}
		prepare.ExpectQuery().WithArgs(args2...).WillReturnRows(rows2)
		prepare.WillBeClosed()
		mock.ExpectCommit()

	default:
		return nil, nil, errors.Errorf("%T not supported", models)
	}
	//
	return db, returning, nil
}

// DB_Update sets up our mock database for updating records.
func DB_Update(models interface{}) (*sql.DB, [][]driver.Value, error) {
	var returning [][]driver.Value
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}
	switch m := models.(type) {
	case *Address:
		returning = append(returning, ReturnArgs(1, "modified"))
		// Set up expected query.
		parts := []string{
			"UPDATE addresses SET",
			"\t\tstreet = $1,",
			"\t\tcity = $2,",
			"\t\tstate = $3,",
			"\t\tzip = $4",
			"\tWHERE",
			"\t\tpk = $5",
			"\tRETURNING modified_tmz",
		}
		qu := mock.ExpectQuery(strings.Join(parts, "\n"))
		// Set up expected args and return columns.
		args := []driver.Value{m.Street, m.City, m.State, m.Zip, m.Id}
		rows := sqlmock.NewRows([]string{"modified_tmz"})
		rows.AddRow(returning[0]...)
		qu.WithArgs(args...)
		qu.WillReturnRows(rows)
		qu.RowsWillBeClosed()

	case []*Address:
		returning = append(returning, ReturnArgs(1, "modified"))
		returning = append(returning, ReturnArgs(1, "modified"))
		// Set up expected query.
		parts := []string{
			"UPDATE addresses SET",
			"\t\tstreet = $1,",
			"\t\tcity = $2,",
			"\t\tstate = $3,",
			"\t\tzip = $4",
			"\tWHERE",
			"\t\tpk = $5",
			"\tRETURNING modified_tmz",
		}
		rows1 := sqlmock.NewRows([]string{"modified_tmz"})
		rows1.AddRow(returning[0][0])
		rows2 := sqlmock.NewRows([]string{"modified_tmz"})
		rows2.AddRow(returning[1][0])
		//
		mock.ExpectBegin()
		prepare := mock.ExpectPrepare(strings.Join(parts, "\n"))
		args1 := []driver.Value{m[0].Street, m[0].City, m[0].State, m[0].Zip, m[0].Id}
		prepare.ExpectQuery().WithArgs(args1...).WillReturnRows(rows1)
		args2 := []driver.Value{m[1].Street, m[1].City, m[1].State, m[1].Zip, m[1].Id}
		prepare.ExpectQuery().WithArgs(args2...).WillReturnRows(rows2)
		prepare.WillBeClosed()
		mock.ExpectCommit()

	default:
		return nil, nil, errors.Errorf("%T not supported", models)
	}
	//
	return db, returning, nil
}
