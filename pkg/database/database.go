package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Querier defines methods for reading data from the database
type Querier interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Executer defines methods for writing data to the database
type Executer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Transaction defines methods for managing transactions
type Transaction interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Database struct that embeds Querier, Executer, and Transaction
type Database struct {
	Querier
	Executer
	Transaction
}

// New constructor to initialize a Database instance with sqlx.DB
func New(db *sqlx.DB) *Database {
	return &Database{
		Querier:     NewQuerier(db),
		Executer:    NewExecuter(db),
		Transaction: NewTransactionManager(db),
	}
}
