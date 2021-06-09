package examples

import (
	"database/sql"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Example is a specific example.
type Example int

const (
	ExSimpleMapper Example = iota
	ExTags
	ExNestedStruct
	ExNestedTwice
	ExScalar
	ExScalarSlice
	ExStruct
	ExStructNotFound
)

// Connect creates a sqlmock DB and configures it for the example.
func Connect(e Example) (DB *sql.DB, err error) {
	var mock sqlmock.Sqlmock
	DB, mock, err = sqlmock.New()
	//
	switch e {
	case ExSimpleMapper:
		mock.ExpectQuery("select +").
			WillReturnRows(
				sqlmock.NewRows([]string{"Message", "Number"}).
					AddRow("Hello, World!", 42).
					AddRow("So long!", 100)).
			RowsWillBeClosed()

	case ExTags:
		mock.ExpectQuery("select +").
			WillReturnRows(
				sqlmock.NewRows([]string{"message", "num"}).
					AddRow("Hello, World!", 42).
					AddRow("So long!", 100)).
			RowsWillBeClosed()

	case ExNestedStruct:
		ts := time.Now().Add(1 * time.Hour)
		mock.ExpectQuery("select +").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "created", "modified", "message", "num"}).
					AddRow(1, ts.Add(1*time.Second), ts.Add(100*time.Second), "Hello, World!", 42).
					AddRow(2, ts.Add(2*time.Second), ts.Add(200*time.Second), "So long!", 100)).
			RowsWillBeClosed()

	case ExNestedTwice:
		ts := time.Now().Add(1 * time.Hour)
		mock.ExpectQuery("select +").
			WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "created", "modified",
					"customer_id", "customer_first", "customer_last",
					"contact_id", "contact_first", "contact_last"}).
					AddRow(1, ts.Add(1*time.Second), ts.Add(100*time.Second), 10, "Bob", "Smith", 100, "Sally", "Johnson").
					AddRow(2, ts.Add(2*time.Second), ts.Add(200*time.Second), 20, "Fred", "Jones", 200, "Betty", "Walker")).
			RowsWillBeClosed()

	case ExScalar:
		mock.ExpectQuery("select +").
			WillReturnRows(sqlmock.NewRows([]string{"n"}).AddRow(64)).RowsWillBeClosed()

	case ExScalarSlice:
		mock.ExpectQuery("select +").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3)).RowsWillBeClosed()

	case ExStruct:
		mock.ExpectQuery("select +").
			WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).
				AddRow(
					time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC))).RowsWillBeClosed()

	case ExStructNotFound:
		mock.ExpectQuery("select +").
			WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "created", "modified",
					"customer_id", "customer_first", "customer_last",
					"contact_id", "contact_first", "contact_last"})).
			RowsWillBeClosed()

	}

	return DB, err
}
