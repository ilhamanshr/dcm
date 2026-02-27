package queries

import (
	"database/sql"
	"errors"
)

func (q *Queries) GetTx() (*sql.Tx, error) {
	tx, ok := q.db.(*sql.Tx)
	if !ok {
		return nil, errors.New("Cannot end tx since queries's DB not a sql.Tx instance")
	}

	if tx == nil {
		return nil, errors.New("Nil DB tx")
	}

	return tx, nil
}
