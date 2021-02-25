package sqlh_test

// NB:  The commented code blocks in here are from benchmarking scanner against sqlx using postgres.
//		Switched to go-sqlmock to have a more predictable database implementation for benchmarks
//		I consider to be more fair.
//		Leaving the commented stuff for now because I may come back to it, possibly as a subpackage.

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/jmoiron/sqlx"
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

// TODO RM
// // DsnFromEnvironment creates a database connection DSN from environment variables.
// func DsnFromEnvironment() (string, error) {
// 	type E struct {
// 		Value  string
// 		Name   string
// 		Driver string
// 	}
// 	vars := []E{
// 		{Value: "", Name: "TESTDB_HOST", Driver: "host"},
// 		{Value: "", Name: "TESTDB_PORT", Driver: "port"},
// 		{Value: "", Name: "TESTDB_SSL", Driver: "sslmode"},
// 		{Value: "", Name: "TESTDB_DB", Driver: "dbname"},
// 		{Value: "", Name: "TESTDB_USER", Driver: "user"},
// 		{Value: "", Name: "TESTDB_PASSWORD", Driver: "password"},
// 	}
// 	var ok bool
// 	dsn := ""
// 	for k := range vars {
// 		if vars[k].Value, ok = os.LookupEnv(vars[k].Name); !ok {
// 			return "", errors.Errorf("Missing environment variable %v", vars[k].Name)
// 		}
// 		dsn = dsn + " " + fmt.Sprintf("%v=%v", vars[k].Driver, vars[k].Value)
// 	}
// 	return dsn, nil
// }

// TODO RM
// InitData initializes the database data for the test.
// func InitData(db interface{}) error {
// 	createTable := `
// 		create table if not exists sqlh_test_json (
// 			id int not null,
// 			created_time timestamp (0) without time zone,
// 			modified_time timestamp (0) without time zone,
// 			price int not null,
// 			quantity int not null,
// 			total int not null,
// 			customer_id int not null,
// 			customer_first character varying not null,
// 			customer_last character varying not null,
// 			vendor_id int not null,
// 			vendor_name character varying not null,
// 			vendor_description character varying not null,
// 			vendor_contact_id int not null,
// 			vendor_contact_first character varying not null,
// 			vendor_contact_last character varying not null
// 		)
// 	`
// 	truncateTable := "truncate table sqlh_test_json cascade"
// 	insert := `
// 		insert into sqlh_test_json (
// 			id, created_time, modified_time,
// 			price, quantity, total,
// 			customer_id, customer_first, customer_last,
// 			vendor_id, vendor_name, vendor_description,
// 			vendor_contact_id, vendor_contact_first, vendor_contact_last
// 		) values (
// 			$1, $2, $3,
// 			$4, $5, $6,
// 			$7, $8, $9,
// 			$10, $11, $12,
// 			$13, $14, $15
// 		)
// 	`
// 	data, err := parseJsonData()
// 	if err != nil {
// 		return err
// 	}
// 	switch d := db.(type) {
// 	case *sql.DB:
// 		if _, err = d.Exec(createTable); err != nil {
// 			return err
// 		} else if _, err = d.Exec(truncateTable); err != nil {
// 			return err
// 		}
// 		stmt, err := d.Prepare(insert)
// 		if err != nil {
// 			return err
// 		}
// 		for _, r := range data {
// 			_, err = stmt.Exec(
// 				r.Id, r.CreatedTime, r.ModifiedTime,
// 				r.Price, r.Quantity, r.Total,
// 				r.CustomerId, r.CustomerFirst, r.CustomerLast,
// 				r.VendorId, r.VendorName, r.VendorDescription,
// 				r.VendorContactId, r.VendorContactFirst, r.VendorContactLast,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 	case *sqlx.DB:
// 		if _, err = d.Exec(createTable); err != nil {
// 			return err
// 		} else if _, err = d.Exec(truncateTable); err != nil {
// 			return err
// 		}
// 		stmt, err := d.Prepare(insert)
// 		if err != nil {
// 			return err
// 		}
// 		for _, r := range data {
// 			_, err = stmt.Exec(
// 				r.Id, r.CreatedTime, r.ModifiedTime,
// 				r.Price, r.Quantity, r.Total,
// 				r.CustomerId, r.CustomerFirst, r.CustomerLast,
// 				r.VendorId, r.VendorName, r.VendorDescription,
// 				r.VendorContactId, r.VendorContactFirst, r.VendorContactLast,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// TODO RM
// Connect connects to the test db using database/sql
// func Connect() (*sql.DB, error) {
// 	dsn, err := DsnFromEnvironment()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sql.Open("postgres", dsn)
// }

// TODO RM
// ConnectSqlx connects to the test db using github.com/jmoiron/sqlx
// func ConnectSqlx() (*sqlx.DB, error) {
// 	dsn, err := DsnFromEnvironment()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sqlx.Open("postgres", dsn)
// }

func TestScanner_Select(t *testing.T) {
	chk := assert.New(t)
	//
	json, err := parseJsonData()
	chk.NoError(err)
	chk.NotNil(json)
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	//
	query := `
	select * from test
	`
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	var dest []*JsonRow
	err = scanner.Select(db, &dest, query)
	chk.NoError(err)
	for k, row := range dest {
		j := json[k]
		chk.Equal(j.Id, row.Id)
		chk.Equal(j.CreatedTime, row.CreatedTime[0:10])
		chk.Equal(j.ModifiedTime, row.ModifiedTime[0:10])

		chk.Equal(j.Price, row.Price)
		chk.Equal(j.Quantity, row.Quantity)
		chk.Equal(j.Total, row.Total)

		chk.Equal(j.CustomerId, row.CustomerId)
		chk.Equal(j.CustomerFirst, row.CustomerFirst)
		chk.Equal(j.CustomerLast, row.CustomerLast)

		chk.Equal(j.VendorId, row.VendorId)
		chk.Equal(j.VendorName, row.VendorName)
		chk.Equal(j.VendorDescription, row.VendorDescription)

		chk.Equal(j.VendorContactId, row.VendorContactId)
		chk.Equal(j.VendorContactFirst, row.VendorContactFirst)
		chk.Equal(j.VendorContactLast, row.VendorContactLast)
	}
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}

func TestScany_Select(t *testing.T) {
	chk := assert.New(t)
	//
	json, err := parseJsonData()
	chk.NoError(err)
	chk.NotNil(json)
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	//
	query := `
	select * from test
	`
	ctx := context.Background()
	var dest []*JsonRow
	err = sqlscan.Select(ctx, db, &dest, query)
	chk.NoError(err)
	for k, row := range dest {
		j := json[k]
		chk.Equal(j.Id, row.Id)
		chk.Equal(j.CreatedTime, row.CreatedTime[0:10])
		chk.Equal(j.ModifiedTime, row.ModifiedTime[0:10])

		chk.Equal(j.Price, row.Price)
		chk.Equal(j.Quantity, row.Quantity)
		chk.Equal(j.Total, row.Total)

		chk.Equal(j.CustomerId, row.CustomerId)
		chk.Equal(j.CustomerFirst, row.CustomerFirst)
		chk.Equal(j.CustomerLast, row.CustomerLast)

		chk.Equal(j.VendorId, row.VendorId)
		chk.Equal(j.VendorName, row.VendorName)
		chk.Equal(j.VendorDescription, row.VendorDescription)

		chk.Equal(j.VendorContactId, row.VendorContactId)
		chk.Equal(j.VendorContactFirst, row.VendorContactFirst)
		chk.Equal(j.VendorContactLast, row.VendorContactLast)
	}
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}
func TestScanner_CodeCoverageNils(t *testing.T) {
	chk := assert.New(t)
	//
	dataRows := sqlmock.NewRows([]string{"A", "B"})
	dataRows.AddRow(1, 2)
	dataRows.AddRow(3, 4)
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	type Dest struct {
		A int
		B int
	}
	var dest []*Dest
	var scanner *sqlh.Scanner
	//
	{ // Select()
		// nil receiver
		scanner = nil
		err := scanner.Select(nil, &dest, "")
		chk.Error(err)
		// nil query
		scanner = &sqlh.Scanner{}
		err = scanner.Select(nil, &dest, "")
		chk.Error(err)
		// nil mapper but non-nil query
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = scanner.Select(db, &dest, "select * from test")
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // ScanRows()
		// nil rows
		err = scanner.ScanRows(nil, &dest)
		chk.Error(err)
		// nil receiver
		scanner = nil
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
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	err = scanner.Select(db, nil, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
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
	dataRows := sqlmock.NewRows([]string{"A", "B"})
	dataRows.AddRow(1, 2)
	dataRows.AddRow(3, 4)
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
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	var dest []*Dest
	err = scanner.Select(db, dest, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
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
func TestScanner_SelectQueryError(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	// query returns error
	mock.ExpectQuery("select (.+)").WillReturnError(errors.Errorf("oops"))
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	var dest []*JsonRow
	err = scanner.Select(db, &dest, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}

func TestScanner_RowError(t *testing.T) {
	chk := assert.New(t)
	//
	json, err := parseJsonData()
	chk.NoError(err)
	chk.NotNil(json)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	j := json[0]
	dataRows.AddRow(
		j.Id, j.CreatedTime, j.ModifiedTime,
		j.Price, j.Quantity, j.Total,
		j.CustomerId, j.CustomerFirst, j.CustomerLast,
		j.VendorId, j.VendorName, j.VendorDescription,
		j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
	)
	j = json[1]
	dataRows.AddRow(
		j.Id, j.CreatedTime, j.ModifiedTime,
		j.Price, j.Quantity, j.Total,
		j.CustomerId, j.CustomerFirst, j.CustomerLast,
		j.VendorId, j.VendorName, j.VendorDescription,
		j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
	)
	dataRows.RowError(1, errors.Errorf("row error"))
	//
	//
	mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	var dest []*JsonRow
	err = scanner.Select(db, &dest, "select * from test")
	chk.Error(err)
	err = mock.ExpectationsWereMet()
	chk.NoError(err)
}

func TestSqlx_Cant(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	{ // works
		type Common struct {
			Id int `db:"id"`
		}
		type NameAge struct {
			Name string `db:"name"`
			Age  int    `db:"age"`
		}
		type Person struct {
			Common
			NameAge
		}
		type Signup struct {
			When string
			Person
		}
		dataRows := sqlmock.NewRows([]string{"when", "id", "name", "age"})
		dataRows.AddRow("Friday", 1, "Bob", 20)
		dbx := sqlx.NewDb(db, "postgres")
		var signupDest Signup
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = dbx.Get(&signupDest, "select * from test")
		chk.NoError(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // nope
		type Common struct {
			Id int `db:"id"`
		}
		type NameAge struct {
			Name string `db:"name"`
			Age  int    `db:"age"`
		}
		type Person struct {
			Common
			NameAge
		}
		type Signup struct {
			When   string
			Person `db:"person"` // <-- sqlx doesn't understand that
		}
		dataRows := sqlmock.NewRows([]string{"when", "person_id", "person_name", "person_age"})
		dataRows.AddRow("Friday", 1, "Bob", 20)
		dbx := sqlx.NewDb(db, "postgres")
		var signupDest Signup
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		err = dbx.Get(&signupDest, "select * from test")
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
}

// TODO RM
// func BenchmarkScannerSelect(b *testing.B) {
// 	chk := assert.New(b)
// 	//
// 	conn, err := Connect()
// 	chk.NoError(err)
// 	chk.NotNil(conn)
// 	defer conn.Close()
// 	//
// 	err = InitData(conn)
// 	chk.NoError(err)
// 	//
// 	query := `
// 	select * from sqlh_test_json order by id asc
// 	`
// 	scanner := &sqlh.Scanner{
// 		Mapper: &set.Mapper{
// 			Tags: []string{"json"},
// 		},
// 	}
// 	//
// 	b.ResetTimer()
// 	//
// 	for k := 0; k < b.N; k++ {
// 		b.StopTimer()
// 		var dest []*JsonRow
// 		rows, err := conn.Query(query)
// 		if err != nil {
// 			b.Errorf(err.Error())
// 			b.FailNow()
// 		}
// 		b.StartTimer()
// 		if err = scanner.ScanRows(rows, &dest); err != nil {
// 			b.Errorf(err.Error())
// 			b.FailNow()
// 		}
// 	}
// }

// TODO RM
// func BenchmarkSqlx(b *testing.B) {
// 	chk := assert.New(b)
// 	//
// 	conn, err := ConnectSqlx()
// 	chk.NoError(err)
// 	chk.NotNil(conn)
// 	defer conn.Close()
// 	//
// 	err = InitData(conn)
// 	chk.NoError(err)
// 	//
// 	query := `
// 	select * from sqlh_test_json order by id asc
// 	`
// 	//
// 	b.ResetTimer()
// 	//
// 	for k := 0; k < b.N; k++ {
// 		b.StopTimer()
// 		var dest []*JsonRow
// 		rows, err := conn.Queryx(query)
// 		if err != nil {
// 			b.Errorf(err.Error())
// 			b.FailNow()
// 		}
// 		defer rows.Close()
// 		b.StartTimer()
// 		for rows.Next() {
// 			record := new(JsonRow)
// 			if err = rows.StructScan(record); err != nil {
// 				b.Errorf(err.Error())
// 				b.FailNow()
// 			}
// 			dest = append(dest, record)
// 		}
// 	}
// }

func BenchmarkSqlMockBaseline(b *testing.B) {
	json, err := parseJsonData()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	//
	b.ResetTimer()
	//
	d := &JsonRow{}
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		rows, err := db.Query("select * from test")
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		b.StartTimer()
		for rows.Next() {
			err = rows.Scan(
				&d.Id, &d.CreatedTime, &d.ModifiedTime,
				&d.Price, &d.Quantity, &d.Total,
				&d.CustomerId, &d.CustomerFirst, &d.CustomerLast,
				&d.VendorId, &d.VendorName, &d.VendorDescription,
				&d.VendorContactId, &d.VendorContactFirst, &d.VendorContactLast,
			)
			if err != nil {
				b.Error(err.Error())
				b.FailNow()
			}
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		b.StopTimer()
		rows.Close()
		b.StartTimer()
	}
}

func BenchmarkSqlMockSqlx(b *testing.B) {
	json, err := parseJsonData()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dbx := sqlx.NewDb(db, "postgres")
	//
	b.ResetTimer()
	//
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		b.StartTimer()
		var dest []*JsonRow
		err = dbx.Select(&dest, "select * from test")
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
	}
}

func BenchmarkSqlMockScany(b *testing.B) {
	json, err := parseJsonData()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	ctx := context.Background()
	//
	b.ResetTimer()
	//
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		b.StartTimer()
		var dest []*JsonRow
		err = sqlscan.Select(ctx, db, &dest, "select * from test")
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
	}
}
func BenchmarkSqlMockScanner(b *testing.B) {
	json, err := parseJsonData()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"json"},
		},
	}
	//
	b.ResetTimer()
	//
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		b.StartTimer()
		var dest []*JsonRow
		err = scanner.Select(db, &dest, "select * from test")
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
	}
}

func BenchmarkSqlMockScannerComplicated(b *testing.B) {
	json, err := parseJsonData()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	dataRows := sqlmock.NewRows([]string{
		"id", "created_time", "modified_time",
		"price", "quantity", "total",
		"customer_id", "customer_first", "customer_last",
		"vendor_id", "vendor_name", "vendor_description",
		"vendor_contact_id", "vendor_contact_first", "vendor_contact_last",
	})
	for _, j := range json {
		dataRows.AddRow(
			j.Id, j.CreatedTime, j.ModifiedTime,
			j.Price, j.Quantity, j.Total,
			j.CustomerId, j.CustomerFirst, j.CustomerLast,
			j.VendorId, j.VendorName, j.VendorDescription,
			j.VendorContactId, j.VendorContactFirst, j.VendorContactLast,
		)
	}
	//
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Error(err.Error())
		b.FailNow()
	}
	//
	type CommonDb struct {
		Id           int
		CreatedTime  string `json:"created_time"`
		ModifiedTime string `json:"modified_time"`
	}
	type Person struct {
		*CommonDb
		First string
		Last  string
	}
	type Vendor struct {
		*CommonDb
		Name        string
		Description string
		Contact     Person
	}
	type Record struct {
		*CommonDb
		Price    int
		Quantity int
		Total    int
		Customer *Person
		Vendor   *Vendor
	}
	scanner := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Elevated:  set.NewTypeList(CommonDb{}),
			Tags:      []string{"json"},
			Join:      "_",
			Transform: strings.ToLower,
		},
	}
	//
	b.ResetTimer()
	//
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		mock.ExpectQuery("select (.+)").WillReturnRows(dataRows)
		b.StartTimer()
		var dest []*Record
		err = scanner.Select(db, &dest, "select * from test")
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
	}
}
