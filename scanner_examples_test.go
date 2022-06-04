package sqlh_test

import (
	"fmt"
	"time"

	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/nofeaturesonlybugs/sqlh/examples"
)

func ExampleScanner_Select_structSlice() {
	type MyStruct struct {
		Message string
		Number  int
	}
	db, err := examples.Connect(examples.ExSimpleMapper)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	scanner := &sqlh.Scanner{
		// Mapper is pure defaults.  Uses exported struct names as column names.
		Mapper: &set.Mapper{},
	}

	fmt.Println("Dest is slice struct:")
	var rv []MyStruct
	err = scanner.Select(db, &rv, "select * from mytable")
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, row := range rv {
		fmt.Printf("%v %v\n", row.Message, row.Number)
	}

	// Output: Dest is slice struct:
	// Hello, World! 42
	// So long! 100
}

func ExampleScanner_Select_structSliceOfPointers() {
	type MyStruct struct {
		Message string
		Number  int
	}
	db, err := examples.Connect(examples.ExSimpleMapper)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	scanner := &sqlh.Scanner{
		// Mapper is pure defaults.  Uses exported struct names as column names.
		Mapper: &set.Mapper{},
	}

	fmt.Println("Dest is slice of pointers:")
	var rv []*MyStruct
	err = scanner.Select(db, &rv, "select * from mytable")
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, row := range rv {
		fmt.Printf("%v %v\n", row.Message, row.Number)
	}

	// Output: Dest is slice of pointers:
	// Hello, World! 42
	// So long! 100
}

func ExampleScanner_Select_tags() {
	type MyStruct struct {
		Message string `json:"message"`
		Number  int    `json:"value" db:"num"`
	}
	db, err := examples.Connect(examples.ExTags)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	scanner := &sqlh.Scanner{
		// Mapper uses struct tag db or json, db higher priority
		Mapper: &set.Mapper{
			Tags: []string{"db", "json"},
		},
	}
	var rv []*MyStruct
	err = scanner.Select(db, &rv, "select message, num from mytable")
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, row := range rv {
		fmt.Printf("%v %v\n", row.Message, row.Number)
	}

	// Output: Hello, World! 42
	// So long! 100
}

func ExampleScanner_Select_nestedStruct() {
	type Common struct {
		Id       int       `json:"id"`
		Created  time.Time `json:"created"`
		Modified time.Time `json:"modified"`
	}
	type MyStruct struct {
		// Structs can share the structure in Common.
		// This would work just as well if the embed was not a pointer.
		// Note how set.Mapper.Elevated is set!
		*Common
		Message string `json:"message"`
		Number  int    `json:"value" db:"num"`
	}
	db, err := examples.Connect(examples.ExNestedStruct)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	scanner := &sqlh.Scanner{
		//
		Mapper: &set.Mapper{
			Elevated: set.NewTypeList(Common{}),
			Tags:     []string{"db", "json"},
		},
	}
	var rv []*MyStruct
	err = scanner.Select(db, &rv, "select id, created, modified, message, num from mytable")
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, row := range rv {
		fmt.Printf("%v %v %v\n", row.Id, row.Message, row.Number)
	}

	// Output: 1 Hello, World! 42
	// 2 So long! 100
}

func ExampleScanner_Select_nestedTwice() {
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
		// Mapper uses struct tag db or json, db higher priority.
		// Mapper elevates Common to same level as other fields.
		Mapper: &set.Mapper{
			Elevated: set.NewTypeList(Common{}),
			Join:     "_",
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
	for _, row := range rv {
		fmt.Printf("%v %v.%v %v %v.%v %v\n", row.Id, row.Customer.Id, row.Customer.First, row.Customer.Last, row.Contact.Id, row.Contact.First, row.Contact.Last)
	}

	// Output: 1 10.Bob Smith 100.Sally Johnson
	// 2 20.Fred Jones 200.Betty Walker
}

func ExampleScanner_Select_scalar() {
	db, err := examples.Connect(examples.ExScalar)
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := sqlh.Scanner{
		Mapper: &set.Mapper{},
	}
	var n int
	err = scanner.Select(db, &n, "select count(*) as n from thetable")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(n)
	// Output: 64
}

func ExampleScanner_Select_scalarSlice() {
	db, err := examples.Connect(examples.ExScalarSlice)
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := sqlh.Scanner{
		Mapper: &set.Mapper{},
	}
	fmt.Println("Dest is slice of scalar:")
	var ids []int
	err = scanner.Select(db, &ids, "select id from thetable where col = ?", "some value")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ids)
	// Output: Dest is slice of scalar:
	// [1 2 3]
}

func ExampleScanner_Select_scalarSliceOfPointers() {
	db, err := examples.Connect(examples.ExScalarSlice)
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := sqlh.Scanner{
		Mapper: &set.Mapper{},
	}
	fmt.Println("Dest is slice of pointer-to-scalar:")
	var ptrs []*int
	err = scanner.Select(db, &ptrs, "select id from thetable where col = ?", "some value")
	if err != nil {
		fmt.Println(err.Error())
	}
	var ids []int
	for _, ptr := range ptrs {
		ids = append(ids, *ptr)
	}
	fmt.Println(ids)
	// Output: Dest is slice of pointer-to-scalar:
	// [1 2 3]
}

func ExampleScanner_Select_struct() {
	db, err := examples.Connect(examples.ExStruct)
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"db", "json"},
		},
	}
	type Temp struct {
		Min time.Time `json:"min"`
		Max time.Time `json:"max"`
	}
	var dest *Temp
	err = scanner.Select(db, &dest, "select min(col) as min, max(col) as max from thetable")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(dest.Min.Format(time.RFC3339), dest.Max.Format(time.RFC3339))
	// Output: 1970-01-01T00:00:00Z 2012-01-01T00:00:00Z
}

func ExampleScanner_Select_structNotFound() {
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
	//
	db, err := examples.Connect(examples.ExStructNotFound)
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := sqlh.Scanner{
		Mapper: &set.Mapper{
			Elevated: set.NewTypeList(Common{}),
			Tags:     []string{"db", "json"},
			Join:     "_",
		},
	}
	query := `
		select
			s.id, s.created, s.modified,
			s.customer_id, c.first as customer_first, c.last as customer_last,
			s.vendor_id as contact_id, v.first as contact_first, v.last as contact_last
		from sales s
		inner join customers c on s.customer_id = c.id
		inner join vendors v on s.vendor_id = v.id
	`
	// When destination is a pointer to struct and no rows are found then the dest pointer
	// remains nil and no error is returned.
	var dest *Sale
	err = scanner.Select(db, &dest, query)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("Is nil: %v\n", dest == nil)

	// Output: Is nil: true
}
