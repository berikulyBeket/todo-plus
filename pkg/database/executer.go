package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// executer implements Executer interface
type executer struct {
	db *sqlx.DB
}

// NewExecuter constructor to initialize a new Executer with a db
func NewExecuter(db *sqlx.DB) Executer {
	return &executer{db: db}
}

// Exec execute a write query (INSERT, UPDATE, DELETE)
func (e *executer) Exec(query string, args ...interface{}) (sql.Result, error) {
	return e.db.Exec(query, args...)
}
