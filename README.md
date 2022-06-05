[![Documentation](https://godoc.org/github.com/nofeaturesonlybugs/sqlh?status.svg)](http://godoc.org/github.com/nofeaturesonlybugs/sqlh)
[![Go Report Card](https://goreportcard.com/badge/github.com/nofeaturesonlybugs/sqlh)](https://goreportcard.com/report/github.com/nofeaturesonlybugs/sqlh)
[![Build Status](https://travis-ci.com/nofeaturesonlybugs/sqlh.svg?branch=master)](https://travis-ci.com/nofeaturesonlybugs/sqlh)
[![codecov](https://codecov.io/gh/nofeaturesonlybugs/sqlh/branch/master/graph/badge.svg)](https://codecov.io/gh/nofeaturesonlybugs/sqlh)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`sqlh` aka `SQL Helper`.

## `sqlh.Scanner`

`sqlh.Scanner` is a powerful database result set scanner.

-   Similar to `jmoiron/sqlx` but supports nested Go `structs`.
-   _Should work with any `database/sql` compatible driver._

## `model.Models`

`model.Models` supports `INSERT|UPDATE` on Go `structs` registered as database _models_, where a _model_ is a language type mapped to a database table.

-   Supports Postgres.
-   Supports grammars that use `?` for parameters **and** have a `RETURNING` clause.
    -   Benchmarked with Sqlite 3.35 -- your mileage may vary.

## `sqlh` Design Philosphy

```
Hand Crafted  |                                         |  Can I Have
   Artisinal  | ======================================= |  My Database
         SQL  |     ^                                   |  Back, Please?
                    |
                    +-- sqlh is here.
```

`sqlh` is easy to use because it lives very close to `database/sql`. The primary goal of `sqlh` is to work with and facilitate using `database/sql` without replacing or hijacking it. When using `sqlh` you manage your `*sql.DB` or create `*sql.Tx` as you normally would and pass those as arguments to functions in `sqlh` when scanning or persisting models; `sqlh` then works within the confines of what you gave it.

When accepting arguments that work directly with the database (`*sql.DB` or `*sql.Tx`) `sqlh` accepts them as interfaces. This means `sqlh` may work with other database packages that define their own types as long as they kept a method set similar to `database/sql`.

The implementation for `sqlh` is fairly straight forward. Primarily this is because all the heavy `reflect` work is offloaded to `set`, which is another of my packages @ https://www.github.com/nofeaturesonlybugs/set

`set` exports a flexible `set.Mapper` for mapping Go `structs` to string names such as database columns. A lot of the power and flexibility exposed by `sqlh` is really derived from `set`. I think this gives `sqlh` an advantage over similar database packages because it's very configurable, performs well, and alleviates `sqlh` from getting bogged down in the complexities of `reflect`.

Here are some `sqlh.Scanner` examples:

```go
type MyStruct struct {
    Message string
    Number  int
}
//
db, err := examples.Connect(examples.ExSimpleMapper)
if err != nil {
    fmt.Println(err.Error())
}
//
scanner := &sqlh.Scanner{
    // Mapper is pure defaults.  Uses exported struct names as column names.
    Mapper: &set.Mapper{},
}
var rv []*MyStruct
err = scanner.Select(db, &rv, "select * from mytable")
if err != nil {
    fmt.Println(err.Error())
}
```

```go
type Common struct {
    Id       int       `json:"id"`
    Created  time.Time `json:"created"`
    Modified time.Time `json:"modified"`
}
type Person struct {
    Common
    First string `json:"first"`
    Last  string `json:"last"`
}
// Note here the natural mapping of SQL columns to nested structs.
type Sale struct {
    Common
    // customer_first and customer_last map to Customer.
    Customer Person `json:"customer"`
    // contact_first and contact_last map to Contact.
    Contact Person `json:"contact"`
}
db, err := examples.Connect(examples.ExNestedTwice)
if err != nil {
    fmt.Println(err.Error())
}
//
scanner := &sqlh.Scanner{
    Mapper: &set.Mapper{
      // Mapper elevates Common to same level as other fields.
      Elevated: set.NewTypeList(Common{}),
      // Nested struct fields joined with _
      Join:     "_",
      // Mapper uses struct tag db or json, db higher priority.
      Tags:     []string{"db", "json"},
    },
}
var rv []*Sale
query := `
        select
            s.id, s.created, s.modified,
            s.customer_id, c.first as customer_first, c.last as customer_last,
            s.vendor_id as contact_id, v.first as contact_first, v.last as contact_last
        from sales s
        inner join customers c on s.customer_id = c.id
        inner join vendors v on s.vendor_id = v.id
    `
err = scanner.Select(db, &rv, query)
if err != nil {
    fmt.Println(err.Error())
}
```

## Roadmap

The development of `sqlh` is essentially following my specific pain points when using `database/sql`:

-   ✓ `SELECT ⭢ for rows.Next() ⭢ row.Scan()` : covered by `sqlh.Scanner`.
-   ✓ `INSERT|UPDATE|UPSERT` CRUD statements : covered by `model.Models`.
    -   `UPSERT` currently supports conflict from primary key; conflicts on arbitrary unique indexes not supported.
-   ⭴ `DELETE` CRUD statements : to be covered by `model.Models`.
-   ⭴ `UPSERT` type operations using index information : to be covered by `model.Models`.
-   ⭴ `Find()` or `Filter()` for advanced `WHERE` clauses and model selection.
-   ⭴ Performance enhancements if possible.
-   ⭴ Relationship management -- maybe.

Personally I find `SELECT|INSERT|UPDATE` to be the most painful and tedious with large queries or tables so those are the features I've addressed first.

## `set.Mapper` Tips

When you want `set.Mapper` to treat a nested struct as a single field rather than a struct itself add it to the `TreatAsScalar` member:

-   `TreatAsScalar : set.NewTypeList( sql.NullBool{}, sql.NullString{} )`

When you use a common nested struct to represent fields present in many of your types consider using the `Elevated` member:

```go
type CommonDB struct {
    Id int
    CreatedAt time.Time
    ModifiedAt time.Time
}
type Something struct {
    CommonDB
    Name string
}
```

Without `Elevated` the `set.Mapper` will generate names like:

```
CommonDBId
CommonDBCreatedAt
CommonDBModifiedAt
Name
```

To prevent `CommonDB` from being part of the name add `CommonDB{}` to the `Elevated` member of the mapper, which elevates the nested fields as if they were defined directly in the parent struct:

```go
Elevated : set.NewTypeList( CommonDB{} )
```

Then the generated names will be:

```
Id
CreatedAt
ModifiedAt
Name
```

You can further customize generated names with struct tags:

```go
type CommonDB struct {
    Id int `json:"id"`
    CreatedAt time.Time `json:"created"`
    ModifiedAt time.Time `json:"modified"`
}
type Something struct {
    CommonDB // No tag necessary since this field is Elevated.
    Name string `json:"name"`
}
```

Specify the tag name to use in the `Tags` member, which is a `[]string`:

```go
Tags : []string{"json"}
```

Now generated names will be:

```
id
created
modified
name
```

If you want to use different names for some fields in your database versus your JSON encoding you can specify multiple `Tags`, with tags listed first taking higher priority:

```go
Tags : []string{"db", "json"} // Uses either db or json, db has higher priority.
```

With the above `Tags`, if `CommonDB` is defined as the following:

```go
type CommonDB struct {
    Id int `json:"id" db:"pk"`
    CreatedAt time.Time `json:"created" db:"created_tmz"`
    ModifiedAt time.Time `json:"modified" db:"modified_tmz"`
}
```

Then the mapped names are:

```
pk
created_tmz
modified_tmz
name
```

## Benchmarks

See my sibling package `sqlhbenchmarks` for my methodology, goals, and interpretation of results.

## API Consistency and Breaking Changes

I am making a very concerted effort to break the API as little as possible while adding features or fixing bugs. However this software is currently in a pre-1.0.0 version and breaking changes _are_ allowed under standard semver. As the API approaches a stable 1.0.0 release I will list any such breaking changes here and they will always be signaled by a bump in _minor_ version.

-   0.4.0 ⭢ 0.5.0
    -   `Models.Register` requires models to be registered with pointer values. This was a "soft" requirement in the prior version where methods would return errors down the line; however in v0.5.0 a panic occurs if models are not registered via pointer.
    -   `Models.QueryBinding` is no longer an interface.
    -   Upgrade dependency `set` to v0.5.1; fields in `model/Model` are redefined accordingly:
        -   `V` and `VSlice` are `set.Value` instead of `*set.Value`
        -   `Mapping` is `set.Mapping` instead of `*set.Mapping`
        -   `PreparedMapping` has replaced the previous `BoundMapping` field.
-   0.3.0 ⭢ 0.4.0
    -   `Transact(fn)` was correctly rolling the transaction back if `fn` returned `err != nil`; however
        the error from `fn` and any potential error from the rollback were not returned from `Transact()`.
        This is fixed in `0.4.0` and while technically a bug fix it _also_ changes the behavior of `Transact()`
        to (correctly) return errors as it should have been doing. As this is a potentially breaking change
        in behavior I have bumped the minor version for this patch.
-   0.2.0 ⭢ 0.3.0
    -   `grammar.Default` renamed to `grammar.Sqlite` -- generated SQL is same as previous version.
    -   `grammar.Grammar` is now an interface where methods now return `(*statements.Query, error)`
        where previously only `(*statements.Query)` was returned.
    -   Package grammar no longer has any panics; errors are returned instead (see previous note).
    -   Prior to this release `model.Models` only ran queries that had followup targets
        for Scan() and panicked when such targets did not exist. This release allows for queries
        that do not have any Scan() targets and will switch to calling Exec() instead of Query() or
        QueryRow() when necessary. An implication of this change is that `Models.Insert()` and
        `Models.Update()` no longer panic in the absence of Scan() targets.
