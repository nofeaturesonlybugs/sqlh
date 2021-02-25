[![Documentation](https://godoc.org/github.com/nofeaturesonlybugs/sqlh?status.svg)](http://godoc.org/github.com/nofeaturesonlybugs/sqlh)
[![Go Report Card](https://goreportcard.com/badge/github.com/nofeaturesonlybugs/sqlh)](https://goreportcard.com/report/github.com/nofeaturesonlybugs/sqlh)
[![Build Status](https://travis-ci.com/nofeaturesonlybugs/sqlh.svg?branch=master)](https://travis-ci.com/nofeaturesonlybugs/sqlh)
[![codecov](https://codecov.io/gh/nofeaturesonlybugs/sqlh/branch/master/graph/badge.svg)](https://codecov.io/gh/nofeaturesonlybugs/sqlh)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`sqlh` provides some utility for `database/sql` and other compatible interfaces.

## Interfaces  
`IQuery` and `IRows` are exported interfaces.  `IQuery` is defined such that a `*sql.DB` or `*sqlx.DB` should be compatible as well as the types for database transactions.  `IRows` is compatible with `*sql.Rows` and consequently `*sqlx.Rows`.

## Scanner  
Currently the workhorse in this package is the `Scanner`.  The methods on `Scanner` are helpers to run queries and populate results.  

Mapping result set `columns` to destination fields in `structs` is controlled by a `set.Mapper`.  

`set.Mapper` is very flexible; you can find the `set` package @ https://github.com/nofeaturesonlybugs/set

Here are some example `Scanners`:   
```go
scanner := &sqlh.Scanner{
    Mapper : &set.Mapper{
        // If using any of the sql.Null* types include them in TreatAsScalar.
        TreatAsAsScalar : set.NewTypeList( sql.NullString{}, sql.NullBool{} ),
        // Nested/embedded structs have their name parts joined with "_".
        Join : "_"
        // Define struct tags to use for SQL column names in order of highest preference.
        // This Tags definition means try `db`, then `json`, then fall back to struct field name.
        Tags : []string{ "db", "json" },
    }
}

scanner := &sqlh.Scanner{
    // This mapper doesn't use struct tags.  It uses field names converted to lower case.
    // Nested/embedded structs will have their name parts joined with "" (empty string).
    Mapper : &set.Mapper{
        Transform : strings.ToLower,
    }
}
```

## Selecting rows  
```go
type Dest struct {
  A string
  B int
}
var dest []*Dest
err = scanner.Select(db, &dest, "select A, B from myTable")
```

## Complicated Destination
```go
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
query := `
  select
    id, created_time, modified_time,
    price, quantity, total,
    customer_id, customer_first, customer_last,
    vendor_id, vendor_name, vendor_description,
    vendor_contact_id, vendor_contact_first, vendor_contact_last,
  from myTable
`
var dest []*Record
err = scanner.Select(db, &dest, query)
```

## A Note to Those that Came Before Me
Before continuing I want to be clear I am not belittling or degrading the work performed by others in this area.  It **is** a difficult problem.  It would be hard enough in a language with looser conventions such as `PHP`; however `Go` is strongly typed which means any such solution must dive into `reflect` -- which is a bear in and of itself -- and adds mountains of complexity onto the fundamental problem about to be presented.

## Why not sqlx, scany, or another package?  
They lack flexbility in ther mappings of columns to nested fields.  

`sqlx` only understands struct embedding.  `scany`'s name generation does not appear to be as flexible as `set.Mapper`.  I suspect the limitations in these packages arise either by conscious design choice or by introducing `reflect` too close to the domain of interacting with the database and unnecessarily complicating matters.

Where this package differs is that it is not at all concerned with mapping columns to struct fields.  It's not concerned with mapping at all.  The entire concern of mapping `strings` to `struct fields` has been exported to my reflection package `set`.

It is within `set` that I tackled this problem.  While a primary goal was to use the solution to scan database results that was not and is not the entire scope of the problem.  Often while implementing the `set.Mapper` I thought:  
  * > What if this mapping is for CSV or `map[string]interface{}`?  
  * > What if the `json` and `db` tags are identical?  Can I allow for tag re-use?
  * > What if I want to map the same `struct T` with different rules?  I shouldn't have to redefine `T` as `TOther`.

As a result of tackling the overall problem in a more general sense and not being directly focused on mapping database results to struct fields I think the resulting `set.Mapper` offers a great amount of flexibility and reuse.  

From there it became somewhat trivial to scan columns.

## Benchmarks  
`scanner_test.go` contains the source for these benchmarks.  However the general idea in each benchmark is to load up `sqlmock` with 100 rows, query, and scan the results.  
```
cpu: Intel(R) Core(TM) i7-7700K CPU @ 4.20GHz
BenchmarkSqlMockBaseline-8                 13310            120201 ns/op            256 B/op 2 allocs/op
BenchmarkSqlMockSqlx-8                     10000            190054 ns/op           3827 B/op 43 allocs/op
BenchmarkSqlMockScany-8                    10000            189010 ns/op           3279 B/op 44 allocs/op
BenchmarkSqlMockScanner-8                  10000            191752 ns/op           4041 B/op 47 allocs/op
BenchmarkSqlMockScannerComplicated-8       10000            190432 ns/op           3887 B/op 47 allocs/op
```

* BenchmarkSqlMockBaseline  
  Your standard `for rows.Next()` code; no special frills, magic, or reflection.  
* BenchmarkSqlMockSqlx  
  `sqlx`'s `db.Select( &dest, query )`  
* BenchmarkSqlMockScany  
  `scany`'s `sqlscan.Select(ctx, db, &dest, query)`
* BenchmarkSqlMockScanner  
  `Scanner.Select( db, &dest, query )` from this package.  
* BenchmarkSqlMockScannerComplicated  
  `Scanner.Select( db, &dest, query )` from this package where `dest` is a more complicated hierarchy
