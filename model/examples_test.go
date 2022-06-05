package model_test

import (
	"fmt"
	"time"

	"github.com/nofeaturesonlybugs/set"

	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model"
	"github.com/nofeaturesonlybugs/sqlh/model/examples"
)

func ExampleModels_Register() {
	// This example demonstrates model registration.

	var Models *model.Models = &model.Models{
		// Mapper and its fields control how Go structs are traversed and mapped to
		// database column names.
		Mapper: &set.Mapper{
			Join: "_",
			Tags: []string{"db", "json"},
		},
		Grammar: grammar.Postgres,

		// StructTag defines the tag name to use when inspecting models.
		// StructTag: "", // Blank defaults to "model"
	}

	//
	// Simple model examples
	//

	// tablename=strings
	//	strings.pk    ↣ auto incrementing key
	//	strings.value
	type StringModel struct {
		// This field specifies the table name in the database.
		//   json:"-" tells encoding/json to ignore this field when marshalling
		//   model:"strings" means the table name is "strings" in the database.
		model.TableName `json:"-" model:"strings"`

		// An auto incrementing primary key field.
		//
		// The mapper is configured to use `db` tag before `json` tag;
		// therefore this maps to strings.pk in the database but json
		// marshals as id
		//
		//  `json:"id" db:"pk" model:"key,auto"`
		//                                 ^-- auto incrementing
		//                              ^-- field is the key or part of composite key
		//                 ^-- maps to strings.pk column
		//          ^-- json marshals to id
		Id int `json:"id" db:"pk" model:"key,auto"`

		// json marshals as value
		// maps to database column strings.value
		Value string `json:"value"`
	}

	// tablename=numbers
	//  numbers.pk    ↣ auto incrementing key
	//  numbers.value
	type NumberModel struct {
		// This model does not include the model.TableName embed; the table name
		// must be specified during registration (see below).
		// model.TableName `json:"-" model:"numbers"`

		Id    int `json:"id" db:"pk" model:"key,auto"`
		Value int `json:"value"`
	}

	// tablename=companies
	//  companies.pk        ↣ auto incrementing key
	//  companies.created   ↣ updates on INSERT
	//  companies.modified  ↣ updates on INSERT and UPDATE
	//  companies.name
	type CompanyModel struct {
		Id int `json:"id" db:"pk" model:"key,auto"`

		// Models can have fields that update during INSERT or UPDATE statements.
		//  `json:"created" model:"inserted"`
		//                           ^-- this column updates on insert
		//  `json:"modified" model:"inserted,updated"`
		//                           ^-- this column updates on insert and updates
		CreatedTime  time.Time `json:"created" model:"inserted"`
		ModifiedTime time.Time `json:"modified" model:"inserted,updated"`

		Name int `json:"name"`
	}

	//
	// Model registration
	//  + Models that embed model.TableName do not need to specify the tablename during registration.
	Models.Register(StringModel{})
	Models.Register(NumberModel{}, model.TableName("numbers"))
	Models.Register(CompanyModel{}, model.TableName("companies"))

	fmt.Println("all done")

	// Output: all done
}

func ExampleModels_insert() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressInsert)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasInserted := func(id int, created time.Time, modified time.Time) error {
		if id == 0 || zero.Equal(created) || zero.Equal(modified) {
			return fmt.Errorf("Record not inserted.")
		}
		return nil
	}
	// A "value" record.
	byVal := examples.Address{
		// Id, CreatedTime, ModifiedTime are updated by the database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}
	// A pointer record.
	byPtr := &examples.Address{
		// Id, CreatedTime, ModifiedTime are updated by the database.
		Street: "4321 The Street",
		City:   "Big City",
		State:  "TS",
		Zip:    "56789",
	}

	// Pass the address of the "value" record.
	if err := examples.Models.Insert(db, &byVal); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasInserted(byVal.Id, byVal.CreatedTime, byVal.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	// The pointer record can be passed directly.
	if err := examples.Models.Insert(db, byPtr); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasInserted(byPtr.Id, byPtr.CreatedTime, byPtr.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	fmt.Println("Models inserted.")

	// Output: Models inserted.
}

func ExampleModels_insertSlice() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressInsertSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasInserted := func(id int, created time.Time, modified time.Time) error {
		if id == 0 || zero.Equal(created) || zero.Equal(modified) {
			return fmt.Errorf("Record not inserted.")
		}
		return nil
	}
	// A slice of values.
	values := []examples.Address{
		// Id, CreatedTime, ModifiedTime are updated by the database.
		{
			Street: "1234 The Street",
			City:   "Small City",
			State:  "ST",
			Zip:    "98765",
		},
		{
			Street: "55 Here We Are",
			City:   "Big City",
			State:  "TS",
			Zip:    "56789",
		},
	}
	// A slice of pointers.
	pointers := []*examples.Address{
		// Id, CreatedTime, ModifiedTime are updated by the database.
		{
			Street: "1234 The Street",
			City:   "Small City",
			State:  "ST",
			Zip:    "98765",
		},
		{
			Street: "55 Here We Are",
			City:   "Big City",
			State:  "TS",
			Zip:    "56789",
		},
	}

	// Slices of values can be passed directly.
	if err := examples.Models.Insert(db, values); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range values {
		if err := WasInserted(model.Id, model.CreatedTime, model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}
	// Slices of pointers can be passed directly.
	if err := examples.Models.Insert(db, pointers); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range pointers {
		if err := WasInserted(model.Id, model.CreatedTime, model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}

	fmt.Println("Models inserted.")

	// Output: Models inserted.
}

func ExampleModels_update() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressUpdate)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasUpdated := func(modified time.Time) error {
		if zero.Equal(modified) {
			return fmt.Errorf("Record not updated.")
		}
		return nil
	}
	// A "value" record.
	byVal := examples.Address{
		Id:          42,
		CreatedTime: time.Now().Add(-1 * time.Hour),
		// ModifiedTime is zero value; will be updated by database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}
	// A pointer record.
	byPtr := &examples.Address{
		Id:          42,
		CreatedTime: time.Now().Add(-1 * time.Hour),
		// ModifiedTime is zero value; will be updated by database.
		Street: "4321 The Street",
		City:   "Big City",
		State:  "TS",
		Zip:    "56789",
	}

	// Pass "value" record by address.
	if err := examples.Models.Update(db, &byVal); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasUpdated(byVal.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	// Pass pointer record directly.
	if err := examples.Models.Update(db, byPtr); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasUpdated(byPtr.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	fmt.Printf("Models updated.")

	// Output: Models updated.
}

func ExampleModels_updateSlice() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressUpdateSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasUpdated := func(modified time.Time) error {
		if zero.Equal(modified) {
			return fmt.Errorf("Record not updated.")
		}
		return nil
	}
	// Slice of values.
	values := []examples.Address{
		// ModifiedTime is not set and updated by the database.
		{
			Id:          42,
			CreatedTime: time.Now().Add(-2 * time.Hour),
			Street:      "1234 The Street",
			City:        "Small City",
			State:       "ST",
			Zip:         "98765",
		},
		{
			Id:          62,
			CreatedTime: time.Now().Add(-1 * time.Hour),
			Street:      "55 Here We Are",
			City:        "Big City",
			State:       "TS",
			Zip:         "56789",
		},
	}
	// Slice of pointers.
	pointers := []*examples.Address{
		// ModifiedTime is not set and updated by the database.
		{
			Id:          42,
			CreatedTime: time.Now().Add(-2 * time.Hour),
			Street:      "1234 The Street",
			City:        "Small City",
			State:       "ST",
			Zip:         "98765",
		},
		{
			Id:          62,
			CreatedTime: time.Now().Add(-1 * time.Hour),
			Street:      "55 Here We Are",
			City:        "Big City",
			State:       "TS",
			Zip:         "56789",
		},
	}

	// Slice of values can be passed directly.
	if err := examples.Models.Update(db, values); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range values {
		if err := WasUpdated(model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}
	// Slice of pointers can be passed directly.
	if err := examples.Models.Update(db, pointers); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range pointers {
		if err := WasUpdated(model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}

	fmt.Println("Models updated.")

	// Output: Models updated.
}

func ExampleModels_upsert() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExUpsert)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasUpserted := func(created time.Time, modified time.Time) error {
		if zero.Equal(created) || zero.Equal(modified) {
			return fmt.Errorf("Record not upserted.")
		}
		return nil
	}
	// A "value" record.
	byVal := examples.Upsertable{
		Id:     "some-unique-string",
		String: "Hello, World!",
		Number: 42,
	}
	// A pointer record.
	byPtr := &examples.Upsertable{
		Id:     "other-unique-string",
		String: "Foo, Bar!",
		Number: 100,
	}

	// Pass "value" record by address.
	if err := examples.Models.Upsert(db, &byVal); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasUpserted(byVal.CreatedTime, byVal.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	// Pass pointer record directly.
	if err := examples.Models.Upsert(db, byPtr); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if err := WasUpserted(byPtr.CreatedTime, byPtr.ModifiedTime); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	fmt.Printf("Models upserted.")

	// Output: Models upserted.
}

func ExampleModels_upsertSlice() {
	var zero time.Time
	//
	// Create a mock database.
	db, err := examples.Connect(examples.ExUpsertSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}
	WasUpserted := func(created time.Time, modified time.Time) error {
		if zero.Equal(created) || zero.Equal(modified) {
			return fmt.Errorf("Record not upserted.")
		}
		return nil
	}
	// Slice of values.
	values := []examples.Upsertable{
		{
			Id:     "some-unique-string",
			String: "Hello, World!",
			Number: 42,
		},
		{
			Id:     "other-unique-string",
			String: "Goodbye, World!",
			Number: 10,
		},
	}
	// Slice of pointers.
	pointers := []*examples.Upsertable{
		{
			Id:     "some-unique-string",
			String: "Hello, World!",
			Number: 42,
		},
		{
			Id:     "other-unique-string",
			String: "Goodbye, World!",
			Number: 10,
		},
	}

	// Pass "values" directly.
	if err := examples.Models.Upsert(db, values); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range values {
		if err := WasUpserted(model.CreatedTime, model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}
	// Pass pointers directly.
	if err := examples.Models.Upsert(db, pointers); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	for _, model := range pointers {
		if err := WasUpserted(model.CreatedTime, model.ModifiedTime); err != nil {
			fmt.Println("err", err.Error())
			return
		}
	}

	fmt.Println("Models upserted.")

	// Output: Models upserted.
}

func ExampleModels_relationship() {
	// This single example shows the common cases of INSERT|UPDATE|UPSERT as distinct code blocks.
	// examples.Relationship has a composite key and no auto-updating fields.
	{
		// Demonstrates INSERT of a single model.
		db, err := examples.Connect(examples.ExRelationshipInsert)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// A "value" model.
		value := examples.Relationship{
			LeftId:  1,
			RightId: 10,
			Toggle:  false,
		}
		// Pass "value" model by address.
		if err = examples.Models.Insert(db, &value); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Insert success.")
	}
	{
		// Demonstrates UPDATE of a single model.
		db, err := examples.Connect(examples.ExRelationshipUpdate)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// A pointer model.
		relate := &examples.Relationship{
			LeftId:  1,
			RightId: 10,
			Toggle:  true,
		}
		// Pass pointer model directly.
		if err = examples.Models.Update(db, relate); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Update success.")
	}
	{
		// Demonstrates UPSERT of a single model.
		db, err := examples.Connect(examples.ExRelationshipUpsert)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//
		relate := &examples.Relationship{
			LeftId:  1,
			RightId: 10,
			Toggle:  false,
		}
		if err = examples.Models.Upsert(db, relate); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Upsert success.")
	}

	// Output: Insert success.
	// Update success.
	// Upsert success.
}

func ExampleModels_relationshipSlice() {
	// This single example shows the common cases of INSERT|UPDATE|UPSERT as distinct code blocks.
	// examples.Relationship has a composite key and no auto-updating fields.
	{
		// Demonstrates INSERT of a slice of models.
		db, err := examples.Connect(examples.ExRelationshipInsertSlice)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// Slice of "values".
		relate := []examples.Relationship{
			{
				LeftId:  1,
				RightId: 10,
				Toggle:  false,
			},
			{
				LeftId:  2,
				RightId: 20,
				Toggle:  true,
			},
			{
				LeftId:  3,
				RightId: 30,
				Toggle:  false,
			},
		}
		// Pass slice of "values" directly.
		if err = examples.Models.Insert(db, relate); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Insert success.")
	}
	{
		// Demonstrates UPDATE of a slice of models.
		db, err := examples.Connect(examples.ExRelationshipUpdateSlice)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// Slice of pointers.
		relate := []*examples.Relationship{
			{
				LeftId:  1,
				RightId: 10,
				Toggle:  true,
			},
			{
				LeftId:  2,
				RightId: 20,
				Toggle:  false,
			},
			{
				LeftId:  3,
				RightId: 30,
				Toggle:  true,
			},
		}
		// Pass slice of pointers directly.
		if err = examples.Models.Update(db, relate); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Update success.")
	}
	{
		// Demonstrates UPSERT of a slice of models.
		db, err := examples.Connect(examples.ExRelationshipUpsertSlice)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//
		relate := []*examples.Relationship{
			{
				LeftId:  1,
				RightId: 10,
				Toggle:  false,
			},
			{
				LeftId:  2,
				RightId: 20,
				Toggle:  true,
			},
			{
				LeftId:  3,
				RightId: 30,
				Toggle:  false,
			},
		}
		if err = examples.Models.Upsert(db, relate); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Upsert success.")
	}

	// Output: Insert success.
	// Update success.
	// Upsert success.
}
