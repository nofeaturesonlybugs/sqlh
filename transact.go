package sqlh

import (
	"database/sql"

	"github.com/nofeaturesonlybugs/errors"
)

// Transact runs fn inside a transaction if Q supports transactions; otherwise it just calls fn(Q).  If a transaction
// is started and fn returns a non-nil error then the transaction is rolled back.
func Transact(Q IQueries, fn func(Q IQueries) error) error {
	var B IBegins
	var T *sql.Tx
	var ok bool
	var err, txnErr error
	if B, ok = Q.(IBegins); !ok {
		return fn(Q)
	} else if T, err = B.Begin(); err != nil {
		return errors.Go(err)
	} else if err = fn(T); err != nil {
		err = errors.Go(err)
		if txnErr = T.Rollback(); txnErr != nil {
			err.(errors.Error).Tag("transaction-rollback", txnErr.Error())
		}
		return err
	} else if err = T.Commit(); err != nil {
		return errors.Go(err)
	}
	return nil
}

// TransactRollback is similar to Transact except the created transaction will always be rolled back; consider using this
// during tests when you do not want to persist changes to the database.
//
// Unlike Transact the database object passed to this function must be of type IBegins so the caller is guaranteed
// fn occurs under a transaction that will be rolled back.
func TransactRollback(B IBegins, fn func(Q IQueries) error) error {
	var T *sql.Tx
	var err error
	if T, err = B.Begin(); err != nil {
		return errors.Go(err)
	} else if err = fn(T); err != nil {
		err = errors.Go(err)
	}
	if e2 := T.Rollback(); e2 != nil {
		if err == nil {
			err = errors.Go(e2)
		} else {
			err.(errors.Error).Tag("transaction-rollback", e2.Error())
		}
	}
	return err
}
