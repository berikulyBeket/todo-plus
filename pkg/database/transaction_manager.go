package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// transactionManager implements Transaction interface
type transactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager constructor to initialize a new TransactionManager with a db
func NewTransactionManager(db *sqlx.DB) *transactionManager {
	return &transactionManager{db: db}
}

// BeginTx start a new transaction and return sqlx.Tx directly
func (t *transactionManager) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return t.db.BeginTxx(ctx, opts)
}
