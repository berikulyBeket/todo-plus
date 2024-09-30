package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres holds the SQLX DB connection pool
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	DB *sqlx.DB
}

// New initialize the Postgres connection using SQLX and raw SQL queries
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	var err error
	for pg.connAttempts > 0 {
		pg.DB, err = sqlx.Open("postgres", url)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), pg.connTimeout)
			defer cancel()

			err = pg.DB.PingContext(ctx)
			if err == nil {
				break
			}
		}

		fmt.Println(err.Error())
		fmt.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL after multiple attempts: %w", err)
	}

	pg.DB.SetMaxOpenConns(pg.maxPoolSize)

	return pg, nil
}

// Close close the database connection
func (p *Postgres) Close() {
	if p.DB != nil {
		p.DB.Close()
	}
}
