package examples_test

import (
	"testing"

	"github.com/nofeaturesonlybugs/sqlh/model/examples"
	"github.com/stretchr/testify/assert"
)

func TestPackage(t *testing.T) {
	chk := assert.New(t)
	//
	m, ms := &examples.Address{}, []*examples.Address{{}, {}}
	type P struct{}
	// inserts
	db, returning, err := examples.DB_Insert(m)
	chk.NoError(err)
	chk.NotEmpty(returning)
	chk.NotNil(db)
	db, returning, err = examples.DB_Insert(ms)
	chk.NoError(err)
	chk.NotEmpty(returning)
	chk.NotNil(db)
	// updates
	db, returning, err = examples.DB_Update(m)
	chk.NoError(err)
	chk.NotEmpty(returning)
	chk.NotNil(db)
	db, returning, err = examples.DB_Update(ms)
	chk.NoError(err)
	chk.NotEmpty(returning)
	chk.NotNil(db)
	// unregistered
	db, returning, err = examples.DB_Insert(P{})
	chk.Error(err)
	chk.Empty(returning)
	chk.Nil(db)
	db, returning, err = examples.DB_Update(P{})
	chk.Error(err)
	chk.Empty(returning)
	chk.Nil(db)
	// new models
	mdb := examples.NewModels()
	chk.NotNil(mdb)
}
