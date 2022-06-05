package examples

import (
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// SentinalTime is a set time value used to generate times.
var SentinalTime time.Time = time.Date(2006, 1, 2, 3, 4, 5, 7, time.Local)

// TimeGenerator uses SentinalTime to return deterministic time values.
type TimeGenerator struct {
	n int
}

// Next returns the next time.Time.
func (tg *TimeGenerator) Next() time.Time {
	rv := SentinalTime.Add(time.Duration(tg.n) * time.Hour)
	tg.n++
	return rv
}

// Example is a specific example.
type Example int

const (
	ExNone Example = iota
	ExAddressInsert
	ExAddressInsertSlice
	ExAddressUpdate
	ExAddressUpdateSlice
	ExAddressSave
	ExAddressSaveSlice
	ExLogEntrySave
	ExRelationshipInsert
	ExRelationshipInsertSlice
	ExRelationshipUpdate
	ExRelationshipUpdateSlice
	ExRelationshipUpsert
	ExRelationshipUpsertSlice
	ExRelationshipSave
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
		allRows := []*sqlmock.Rows{
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(ReturnArgs(1, "pk", "created", "modified")...),
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(ReturnArgs(1, "pk", "created", "modified")...),
		}
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1234 The Street", "Small City", "ST", "98765").
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("4321 The Street", "Big City", "TS", "56789").
			WillReturnRows(allRows[1]).
			RowsWillBeClosed()

	case ExAddressInsertSlice:
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		for k := 0; k < 2; k++ {
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
		}

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
		allRows := []*sqlmock.Rows{
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(ReturnArgs(1, "modified")...),
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(ReturnArgs(1, "modified")...),
		}
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1234 The Street", "Small City", "ST", "98765", 42).
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("4321 The Street", "Big City", "TS", "56789", 42).
			WillReturnRows(allRows[1]).
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
		for k := 0; k < 2; k++ {
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
		}

	case ExAddressSave:
		var tg TimeGenerator
		var times []time.Time = []time.Time{
			tg.Next(), tg.Next(),
		}

		// The INSERT portion
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		allRows := []*sqlmock.Rows{
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(1, times[0], times[0]),
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(2, times[1], times[1]),
		}
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1234 The Street", "Small City", "ST", "98765").
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("55 Here We Are", "Big City", "TS", "56789").
			WillReturnRows(allRows[1]).
			RowsWillBeClosed()

		// The UPDATE portion
		parts = []string{
			"UPDATE addresses SET",
			"\t\tstreet = $1,",
			"\t\tcity = $2,",
			"\t\tstate = $3,",
			"\t\tzip = $4",
			"\tWHERE",
			"\t\tpk = $5",
			"\tRETURNING modified_tmz",
		}
		allRows = []*sqlmock.Rows{
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(times[0].Add(time.Hour)),
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(times[1].Add(time.Hour)),
		}
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("1 New Street", "Small City", "ST", "99111", 1).
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		mock.ExpectQuery(strings.Join(parts, "\n")).
			WithArgs("2 New Street", "Big City", "TS", "99222", 2).
			WillReturnRows(allRows[1]).
			RowsWillBeClosed()

	case ExAddressSaveSlice:
		var tg TimeGenerator
		var times []time.Time = []time.Time{
			tg.Next(), tg.Next(),
		}

		// The INSERT portion
		parts := []string{
			"INSERT INTO addresses",
			"\t\t( street, city, state, zip )",
			"\tVALUES",
			"\t\t( $1, $2, $3, $4 )",
			"\tRETURNING pk, created_tmz, modified_tmz",
		}
		allRows := []*sqlmock.Rows{
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(1, times[0], times[0]),
			sqlmock.NewRows([]string{"pk", "created_tmz", "modified_tmz"}).
				AddRow(2, times[1], times[1]),
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectQuery().
			WithArgs("1234 The Street", "Small City", "ST", "98765").
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		prepared.ExpectQuery().
			WithArgs("55 Here We Are", "Big City", "TS", "56789").
			WillReturnRows(allRows[1]).
			RowsWillBeClosed()
		prepared.WillBeClosed()
		mock.ExpectCommit()

		// The UPDATE portion
		parts = []string{
			"UPDATE addresses SET",
			"\t\tstreet = $1,",
			"\t\tcity = $2,",
			"\t\tstate = $3,",
			"\t\tzip = $4",
			"\tWHERE",
			"\t\tpk = $5",
			"\tRETURNING modified_tmz",
		}
		allRows = []*sqlmock.Rows{
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(times[0].Add(time.Hour)),
			sqlmock.NewRows([]string{"modified_tmz"}).
				AddRow(times[1].Add(time.Hour)),
		}
		mock.ExpectBegin()
		prepared = mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectQuery().
			WithArgs("1 New Street", "Small City", "ST", "99111", 1).
			WillReturnRows(allRows[0]).
			RowsWillBeClosed()
		prepared.ExpectQuery().
			WithArgs("2 New Street", "Big City", "TS", "99222", 2).
			WillReturnRows(allRows[1]).
			RowsWillBeClosed()
		prepared.WillBeClosed()
		mock.ExpectCommit()

	case ExLogEntrySave:
		parts := []string{
			"INSERT INTO log",
			"\t\t( message )",
			"\tVALUES",
			"\t\t( $1 )",
		}
		mock.ExpectBegin()
		prepared := mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectExec().
			WithArgs("Hello, World!").
			WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().
			WithArgs("Foo, Bar!").
			WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().
			WithArgs("The llamas are escaping!").
			WillReturnResult(sqlmock.NewResult(0, 1))
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

	case ExRelationshipSave:
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
		prepared.ExpectExec().WithArgs(1, 2, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(10, 20, false).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.WillBeClosed()
		mock.ExpectCommit()
		mock.ExpectBegin()
		prepared = mock.ExpectPrepare(strings.Join(parts, "\n"))
		prepared.ExpectExec().WithArgs(1, 2, true).WillReturnResult(sqlmock.NewResult(0, 1))
		prepared.ExpectExec().WithArgs(10, 20, true).WillReturnResult(sqlmock.NewResult(0, 1))
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
		allRows := []*sqlmock.Rows{
			sqlmock.NewRows([]string{"created_tmz", "modified_tmz"}).
				AddRow(ReturnArgs(1, "created", "modified")...),
			sqlmock.NewRows([]string{"created_tmz", "modified_tmz"}).
				AddRow(ReturnArgs(1, "created", "modified")...),
		}
		qu := mock.ExpectQuery(strings.Join(parts, "\n"))
		qu.WithArgs("some-unique-string", "Hello, World!", 42)
		qu.WillReturnRows(allRows[0])
		qu.RowsWillBeClosed()
		qu = mock.ExpectQuery(strings.Join(parts, "\n"))
		qu.WithArgs("other-unique-string", "Foo, Bar!", 100)
		qu.WillReturnRows(allRows[1])
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
		for k := 0; k < 2; k++ {
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

	}
	//
	return DB, err
}
