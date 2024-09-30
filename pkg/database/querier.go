package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// querier implements Querier interface
type querier struct {
	db *sqlx.DB
}

// NewQuerier constructor to initialize a new Querier with a db
func NewQuerier(db *sqlx.DB) Querier {
	return &querier{db: db}
}

// Get fetch a single record from the database
func (q *querier) Get(dest interface{}, query string, args ...interface{}) error {
	return q.db.Get(dest, query, args...)
}

// Select fetch multiple records from the database
func (q *querier) Select(dest interface{}, query string, args ...interface{}) error {
	return q.db.Select(dest, query, args...)
}

// QueryRow executes a query that returns at most one row
func (q *querier) QueryRow(query string, args ...interface{}) *sql.Row {
	return q.db.QueryRow(query, args...)
}

// Query executes a query that returns multiple rows
func (q *querier) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return q.db.Query(query, args...)
}
