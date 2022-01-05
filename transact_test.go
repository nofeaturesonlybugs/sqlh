package sqlh_test

import (
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nofeaturesonlybugs/errors"
	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/sqlh"
)

func TestTransact(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	// Several of the tests use a insert...insert function.
	Insert2xFunc := func(Q sqlh.IQueries) error {
		var err error
		if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
			return err
		} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
			return err
		}
		return nil
	}
	//
	type Test struct {
		Name        string
		MockFn      func(sqlmock.Sqlmock)
		TransactFn  func(sqlh.IQueries) error
		ExpectError bool
	}
	tests := []Test{
		{
			Name: "no errors",
			MockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit()
			},
			TransactFn:  Insert2xFunc,
			ExpectError: false,
		},
		{
			Name: "commit error",
			MockFn: func(sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit().WillReturnError(errors.Errorf("commit error"))
			},
			TransactFn:  Insert2xFunc,
			ExpectError: true,
		},
		{
			// Calling sqlh.Transact() inside sqlh.Transaction should result in no additional calls
			// to Begin().
			Name: "nested",
			MockFn: func(sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("update+").WithArgs("x", "y", "z").WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit()
			},
			TransactFn: func(Q sqlh.IQueries) error {
				var err error
				if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
					return err
				} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
					return err
				} else if err = sqlh.Transact(Q, func(Q sqlh.IQueries) error {
					var err error
					if _, err = Q.Exec("update my table set", "x", "y", "z"); err != nil {
						return err
					}
					return nil
				}); err != nil {
					return err
				}
				return nil
			},
			ExpectError: false,
		},
		{
			// Fn returns error so expect a rollback and error to filter upwards.
			Name: "rollback",
			MockFn: func(sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnError(errors.Errorf("insert error"))
				mock.ExpectRollback()
			},
			TransactFn:  Insert2xFunc,
			ExpectError: true,
		},
		{
			// Fn returns error so expect rollback but now rollback also errors.
			Name: "rollback error",
			MockFn: func(sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
				mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnError(errors.Errorf("insert error"))
				mock.ExpectRollback().WillReturnError(errors.Errorf("rollback error"))
			},
			TransactFn:  Insert2xFunc,
			ExpectError: true,
		},
		{
			Name: "begin error",
			MockFn: func(sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.Errorf("begin error"))
			},
			TransactFn:  Insert2xFunc,
			ExpectError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			test.MockFn(mock)
			err := sqlh.Transact(db, test.TransactFn)
			if test.ExpectError {
				chk.Error(err)
			} else {
				chk.NoError(err)
			}
			err = mock.ExpectationsWereMet()
			chk.NoError(err)
		})
	}
}

func TestTransactWithRollback(t *testing.T) {
	chk := assert.New(t)
	//
	db, mock, err := sqlmock.New()
	chk.NoError(err)
	chk.NotNil(db)
	chk.NotNil(mock)
	//
	{ // begin, insert, insert, rollback
		mock.ExpectBegin()
		mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
		mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
		mock.ExpectRollback()
		//
		err = sqlh.TransactRollback(db, func(Q sqlh.IQueries) error {
			var err error
			if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
				return err
			} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
				return err
			}
			return nil
		})
		chk.NoError(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // begin (with begin error)
		mock.ExpectBegin().WillReturnError(errors.Errorf("begin error"))
		//
		err = sqlh.TransactRollback(db, func(Q sqlh.IQueries) error {
			var err error
			if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
				return err
			} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
				return err
			}
			return nil
		})
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // begin, insert, insert, rollback (with rollback error)
		mock.ExpectBegin()
		mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
		mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
		mock.ExpectRollback().WillReturnError(errors.Errorf("rollback error"))
		//
		err = sqlh.TransactRollback(db, func(Q sqlh.IQueries) error {
			var err error
			if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
				return err
			} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
				return err
			}
			return nil
		})
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
	{ // begin, insert, insert, fn returns error, rollback (with rollback error)
		mock.ExpectBegin()
		mock.ExpectExec("insert+").WithArgs("a", "b", "c").WillReturnResult(driver.ResultNoRows)
		mock.ExpectExec("insert+").WithArgs("1", "2", "3").WillReturnResult(driver.ResultNoRows)
		mock.ExpectRollback().WillReturnError(errors.Errorf("rollback error"))
		//
		err = sqlh.TransactRollback(db, func(Q sqlh.IQueries) error {
			var err error
			if _, err = Q.Exec("insert into my table", "a", "b", "c"); err != nil {
				return err
			} else if _, err = Q.Exec("insert into my table", "1", "2", "3"); err != nil {
				return err
			}
			return errors.Errorf("force error")
		})
		chk.Error(err)
		err = mock.ExpectationsWereMet()
		chk.NoError(err)
	}
}
