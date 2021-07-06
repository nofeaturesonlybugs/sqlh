package model_test

import (
	"fmt"
	"time"

	"github.com/nofeaturesonlybugs/sqlh/model/examples"
)

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
