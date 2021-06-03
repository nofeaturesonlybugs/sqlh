package examples

import (
	"time"

	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh/grammar"
	"github.com/nofeaturesonlybugs/sqlh/model"
)

var (
	// Models is a sample model.Database.
	//
	// Most of the magic occurs in the Mapper (set.Mapper); if your models follow a consistent logic
	// for struct-tags to column-names then you may only need a single model.Database for your application.
	//
	// However creating a model.Database is relatively easy and there's nothing to stop you from creating
	// multiple of them with different Mapper (set.Mapper) to handle inconsistency in your models.
	Models = &model.Models{
		// Mapper defines how struct fields map to friendly column names (i.e. database names).
		Mapper: &set.Mapper{
			Join: "_",
			Tags: []string{"db", "json"},
		},
		// This instance uses a Postgres grammar.
		Grammar: grammar.Postgres,
	}
)

// NewModels returns a model.Models type.
func NewModels() *model.Models {
	rv := &model.Models{
		Mapper: &set.Mapper{
			Join: "_",
			Tags: []string{"db", "json"},
		},
		Grammar: grammar.Postgres,
	}
	rv.Register(&Address{})
	rv.Register(&Person{})
	rv.Register(&PersonAddress{})
	return rv
}

func init() {
	// Somewhere in your application you need to register all types to be used as models.
	Models.Register(&Address{})
	Models.Register(&Person{})
	Models.Register(&PersonAddress{})
}

// Address is a simple model representing an address.
type Address struct {
	// This member tells the model.Database the name of the table for this model; think of it
	// like xml.Name when using encoding/xml.
	model.TableName `json:"-" model:"addresses"`
	//
	// The struct fields and column names.  The example Mapper allows the db and json
	// to share names where they are the same; however db is higher in precedence than
	// json so where present that is the database column name.
	Id           int       `json:"id" db:"pk" model:"key,auto"`
	CreatedTime  time.Time `json:"created_time" db:"created_tmz" model:"inserted"`
	ModifiedTime time.Time `json:"modified_time" db:"modified_tmz" model:"inserted,updated"`
	Street       string    `json:"street"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Zip          string    `json:"zip"`
}

// Person is a simple model representing a person.
type Person struct {
	model.TableName `json:"-" model:"people"`
	//
	Id           int       `json:"id" db:"pk" model:"key,auto"`
	SpouseId     int       `json:"spouse_id" db:"spouse_fk" model:"foreign"`
	CreatedTime  time.Time `json:"created_time" db:"created_tmz" model:"inserted"`
	ModifiedTime time.Time `json:"modified_time" db:"modified_tmz" model:"inserted,updated"`
	First        string    `json:"first"`
	Last         string    `json:"last"`
	Age          int       `json:"age"`
	SSN          string    `json:"ssn" model:"unique"`
}

// PersonAddress links a person to an address.
type PersonAddress struct {
	model.TableName `json:"-" model:"relate_people_addresses"`
	//
	PersonId  int `json:"person_id" db:"person_fk" model:"key"`
	AddressId int `json:"address_id" db:"address_fk" model:"key"`
}
