package model_test

import (
	"fmt"
	"time"

	"github.com/nofeaturesonlybugs/sqlh/model/examples"
)

func ExampleModels_insert() {
	model := &examples.Address{
		// Id, CreatedTime, ModifiedTime are not set and updated by the database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}

	// Create a mock database.  Unlike a production database the mock DB tells us
	// which values to expect when scanning back into models via returning.
	db, returning, err := examples.DB_Insert(model)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Insert our model.  The values from returning (see above) are scanned into
	// auto updating fields of the model (pk, created, modified).
	if err := examples.Models.Insert(db, model); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check values from returning were scanned into the model.
	if model.Id != returning[0][0].(int) {
		fmt.Println("unexpected id")
	}
	if !model.CreatedTime.Equal(returning[0][1].(time.Time)) {
		fmt.Println("unexpected created time")
	}
	if !model.ModifiedTime.Equal(returning[0][2].(time.Time)) {
		fmt.Println("unexpected modified time")
	}
	fmt.Printf("Model inserted.")

	// Output: Model inserted.
}

func ExampleModels_insertSlice() {
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

	// Create a mock database.  Unlike a production database the mock DB tells us
	// which values to expect when scanning back into models via returning.
	db, returning, err := examples.DB_Insert(models)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Insert our model.  The values from returning (see above) are scanned into
	// auto updating fields of the model (pk, created, modified).
	if err := examples.Models.Insert(db, models); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check values from returning were scanned into the model.
	for k, model := range models {
		if model.Id != returning[k][0].(int) {
			fmt.Println("unexpected id")
			return
		}
		if model.CreatedTime != returning[k][1].(time.Time) {
			fmt.Println("unexpected created time")
			return
		}
		if model.ModifiedTime != returning[k][2].(time.Time) {
			fmt.Println("unexpected modified time")
			return
		}
	}
	fmt.Printf("%v model(s) updated.", len(models))
}

func ExampleModels_update() {
	model := &examples.Address{
		Id:          42,
		CreatedTime: time.Now().Add(-1 * time.Hour),
		// ModifiedTime is zero value; will be updated by database.
		Street: "1234 The Street",
		City:   "Small City",
		State:  "ST",
		Zip:    "98765",
	}

	// Create a mock database.  Unlike a production database the mock DB tells us
	// which values to expect when scanning back into models via returning.
	db, returning, err := examples.DB_Update(model)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Update our model.  The values from returning (see above) are scanned into
	// auto updating fields of the model (modified).
	if err := examples.Models.Update(db, model); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check values from returning were scanned into the model.
	if !model.ModifiedTime.Equal(returning[0][0].(time.Time)) {
		fmt.Println("unexpected modified time")
	}
	fmt.Printf("Model updated.")

	// Output: Model updated.
}

func ExampleModels_updateSlice() {
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

	// Create a mock database.  Unlike a production database the mock DB tells us
	// which values to expect when scanning back into models via returning.
	db, returning, err := examples.DB_Update(models)
	if err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Update our models.  The values from returning (see above) are scanned into
	// auto updating fields of the model (modified).
	if err := examples.Models.Update(db, models); err != nil {
		fmt.Println("err", err.Error())
		return
	}

	// Check values from returning were scanned into the model.
	for k, model := range models {
		if model.ModifiedTime != returning[k][0].(time.Time) {
			fmt.Println("unexpected modified time")
			return
		}
	}
	fmt.Printf("%v model(s) updated.", len(models))

	// Output: 2 model(s) updated.
}
