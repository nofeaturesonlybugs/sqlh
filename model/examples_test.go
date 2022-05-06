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
	//  + Models must be registered as pointers or a panic occurs.
	//  + Models that embed model.TableName do not need to specify the tablename during registration.
	Models.Register(&StringModel{})
	Models.Register(&NumberModel{}, model.TableName("numbers"))
	Models.Register(&CompanyModel{}, model.TableName("companies"))

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Registering a non-pointer panics.")
			}
		}()
		Models.Register(StringModel{})
	}()

	fmt.Println("all done")

	// Output: Registering a non-pointer panics.
	// all done
}

func ExampleModels_insert() {
	var zero time.Time
	model := &examples.Address{
		// Id, CreatedTime, ModifiedTime are not set and updated by the database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}

	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressInsert)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Insert our model.
	if err := examples.Models.Insert(db, model); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	if model.Id == 0 {
		fmt.Println("id not updated")
	}
	if model.CreatedTime.Equal(zero) {
		fmt.Println("created time not updated")
	}
	if model.ModifiedTime.Equal(zero) {
		fmt.Println("modified time not updated")
	}
	fmt.Printf("Model inserted.")

	// Output: Model inserted.
}

func ExampleModels_insertSlice() {
	var zero time.Time
	models := []*examples.Address{
		// Id, CreatedTime, ModifiedTime are not set and updated by the database.
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

	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressInsertSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Insert our models.
	if err := examples.Models.Insert(db, models); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	for _, model := range models {
		if model.Id == 0 {
			fmt.Println("id not updated")
			return
		}
		if model.CreatedTime.Equal(zero) {
			fmt.Println("created time not updated")
			return
		}
		if model.ModifiedTime.Equal(zero) {
			fmt.Println("modified time not updated")
			return
		}
	}
	fmt.Printf("%v model(s) inserted.", len(models))

	// Output: 2 model(s) inserted.
}

func ExampleModels_update() {
	var zero time.Time
	model := &examples.Address{
		Id:          42,
		CreatedTime: time.Now().Add(-1 * time.Hour),
		// ModifiedTime is zero value; will be updated by database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}

	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressUpdate)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Update our model.
	if err := examples.Models.Update(db, model); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	if model.ModifiedTime.Equal(zero) {
		fmt.Println("modified time not updated")
	}
	fmt.Printf("Model updated.")

	// Output: Model updated.
}

func ExampleModels_updateSlice() {
	var zero time.Time
	models := []*examples.Address{
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

	// Create a mock database.
	db, err := examples.Connect(examples.ExAddressUpdateSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Update our models.
	if err := examples.Models.Update(db, models); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	for _, model := range models {
		if model.ModifiedTime.Equal(zero) {
			fmt.Println("modified time not updated")
			return
		}
	}
	fmt.Printf("%v model(s) updated.", len(models))

	// Output: 2 model(s) updated.
}

func ExampleModels_upsert() {
	var zero time.Time
	model := &examples.Upsertable{
		Id:     "some-unique-string",
		String: "Hello, World!",
		Number: 42,
	}

	// Create a mock database.
	db, err := examples.Connect(examples.ExUpsert)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Upsert our model.
	if err := examples.Models.Upsert(db, model); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	if model.CreatedTime.Equal(zero) {
		fmt.Println("created time not updated")
	}
	if model.ModifiedTime.Equal(zero) {
		fmt.Println("modified time not updated")
	}
	fmt.Printf("Model upserted.")

	// Output: Model upserted.
}

func ExampleModels_upsertSlice() {
	var zero time.Time
	models := []*examples.Upsertable{
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

	// Create a mock database.
	db, err := examples.Connect(examples.ExUpsertSlice)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Upsert our models.
	if err := examples.Models.Upsert(db, models); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check expected model fields are no longer zero values.
	for _, model := range models {
		if model.CreatedTime.Equal(zero) {
			fmt.Println("created time not updated")
			return
		}
		if model.ModifiedTime.Equal(zero) {
			fmt.Println("modified time not updated")
			return
		}
	}
	fmt.Printf("%v model(s) upserted.", len(models))

	// Output: 2 model(s) upserted.
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
		//
		relate := &examples.Relationship{
			LeftId:  1,
			RightId: 10,
			Toggle:  false,
		}
		if err = examples.Models.Insert(db, relate); err != nil {
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
		//
		relate := &examples.Relationship{
			LeftId:  1,
			RightId: 10,
			Toggle:  true,
		}
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
		//
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
