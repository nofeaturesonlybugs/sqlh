package examples

import (
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Example is a specific example.
type Example int

const (
	ExNone Example = iota
	ExAddressInsert
	ExAddressInsertSlice
	ExAddressUpdate
	ExAddressUpdateSlice
	ExRelationshipInsert
	ExRelationshipInsertSlice
	ExRelationshipUpdate
	ExRelationshipUpdateSlice
	ExRelationshipUpsert
	ExRelationshipUpsertSlice
	ExUpsert
	ExUpsertSlice
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

// Connect creates a sqlmock DB and configures it for the example.
func Connect(e Example) (DB *sql.DB, err error) {
	var mock sqlmock.Sqlmock
	DB, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	//
	switch e {
	case ExAddressInsert:
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		rows := sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
			AddRow(ReturnArgs(1, "pk", "created", "modified")...)
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1234 The Street", "Small City", "ST", "98765").
			WillReturnRows(rows).
			RowsWillBeClosed()

	case ExAddressInsertSlice:
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		rows := sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
			AddRow(ReturnArgs(1, "pk", "created", "modified")...)
		prepared.ExpectQuery().
			WithArgs("1234 The Street", "Small City", "ST", "98765").
			WillReturnRows(rows).
			RowsWillBeClosed()
		rows = sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
			AddRow(ReturnArgs(1, "pk", "created", "modified")...)
		prepared.ExpectQuery().
			WithArgs("55 Here We Are", "Big City", "TS", "56789").
			WillReturnRows(rows).
			RowsWillBeClosed()
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExAddressUpdate:
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
		rows := sqlmock.NewRows([]string{"modified_tmz"}).
			AddRow(ReturnArgs(1, "modified")...)
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1234 The Street", "Small City", "ST", "98765", 42).
			WillReturnRows(rows).
			RowsWillBeClosed()

	case ExAddressUpdateSlice:
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
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		rows := sqlmock.NewRows([]string{"modified_tmz"}).
			AddRow(ReturnArgs(1, "modified")...)
		prepared.ExpectQuery().
			WithArgs("1234 The Street", "Small City", "ST", "98765", 42).
			WillReturnRows(rows).
			RowsWillBeClosed()
		rows = sqlmock.NewRows([]string{"modified_tmz"}).
			AddRow(ReturnArgs(1, "modified")...)
		prepared.ExpectQuery().
			WithArgs("55 Here We Are", "Big City", "TS", "56789", 62).
			WillReturnRows(rows).
			RowsWillBeClosed()
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExRelationshipInsert:
		parts := []string{
			"INSERT INTO relationship",
			"\t\t( left_fk, right_fk, toggle )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
		}
		mock.ExpectExec(strings.Join(parts, "\n")).WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))

	case ExRelationshipInsertSlice:
		parts := []string{
			"INSERT INTO relationship",
			"\t\t( left_fk, right_fk, toggle )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(3, 30, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExRelationshipUpdate:
		parts := []string{
			"UPDATE relationship SET",
			"\t\ttoggle = $1",
			"\tWHERE",
			"\t\tleft_fk = $2 AND right_fk = $3",
		}
		mock.ExpectExec(strings.Join(parts, "\n")).WithArgs(true, 1, 10).WillReturnResult(sqlmock.NewResult(0, 1))

	case ExRelationshipUpdateSlice:
		parts := []string{
			"UPDATE relationship SET",
			"\t\ttoggle = $1",
			"\tWHERE",
			"\t\tleft_fk = $2 AND right_fk = $3",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectExec().WithArgs(true, 1, 10).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(false, 2, 20).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(true, 3, 30).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExRelationshipUpsert:
		parts := []string{
			"INSERT INTO relationship AS dest",
			"\t\t( left_fk, right_fk, toggle )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
			"\tON CONFLICT( left_fk, right_fk ) DO UPDATE SET",
			"\t\ttoggle = EXCLUDED.toggle",
			"\t\tWHERE (",
			"\t\t\tdest.toggle <> EXCLUDED.toggle",
			"\t\t)",
		}
		mock.ExpectExec(strings.Join(parts, "\n")).WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))

	case ExRelationshipUpsertSlice:
		parts := []string{
			"INSERT INTO relationship AS dest",
			"\t\t( left_fk, right_fk, toggle )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
			"\tON CONFLICT( left_fk, right_fk ) DO UPDATE SET",
			"\t\ttoggle = EXCLUDED.toggle",
			"\t\tWHERE (",
			"\t\t\tdest.toggle <> EXCLUDED.toggle",
			"\t\t)",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectExec().WithArgs(1, 10, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(2, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(3, 30, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExUpsert:
		parts := []string{
			"INSERT INTO upsertable AS dest",
			"\t\t( pk, string, number )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
			"\tON CONFLICT( pk ) DO UPDATE SET",
			"\t\tstring = EXCLUDED.string, number = EXCLUDED.number",
			"\t\tWHERE (",
			"\t\t\tdest.string <> EXCLUDED.string OR dest.number <> EXCLUDED.number",
			"\t\t)",
			"\tRETURNING created_tmz, modified_tmz",
		}
		qu := mock.ExpectQuery(strings.Join(parts, "\n"))
		rows := sqlmock.NewRows([]string{"created_tmz", "modified_tmz"})
		rows.AddRow(ReturnArgs(1, "created", "modified")...)
		qu.WithArgs("some-unique-string", "Hello, World!", 42)
		qu.WillReturnRows(rows)
		qu.RowsWillBeClosed()

	case ExUpsertSlice:
		parts := []string{
			"INSERT INTO upsertable AS dest",
			"\t\t( pk, string, number )",
			"\tVALUES",
			"\t\t( $1, $2, $3 )",
			"\tON CONFLICT( pk ) DO UPDATE SET",
			"\t\tstring = EXCLUDED.string, number = EXCLUDED.number",
			"\t\tWHERE (",
			"\t\t\tdest.string <> EXCLUDED.string OR dest.number <> EXCLUDED.number",
			"\t\t)",
			"\tRETURNING created_tmz, modified_tmz",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		rows := sqlmock.NewRows([]string{"created_tmz", "modified_tmz"}).
			AddRow(ReturnArgs(1, "created", "modified")...)
		prepared.ExpectQuery().
			WithArgs("some-unique-string", "Hello, World!", 42).
			WillReturnRows(rows)
		rows = sqlmock.NewRows([]string{"created_tmz", "modified_tmz"}).
			AddRow(ReturnArgs(1, "created", "modified")...)
		prepared.ExpectQuery().
			WithArgs("other-unique-string", "Goodbye, World!", 10).
			WillReturnRows(rows)
		prepared.WillBeClosed()
		mock.ExpectCommit()

	}
	//
	return DB, err
}
