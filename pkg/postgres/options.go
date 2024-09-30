package postgres

import "time"

// Option defines a type for applying various configurations to the Postgres struct
type Option func(*Postgres)

// MaxPoolSize sets the maximum number of connections in the Postgres connection pool
func MaxPoolSize(size int) Option {
	return func(c *Postgres) {
		c.maxPoolSize = size
	}
}

// ConnAttempts sets the number of attempts to reconnect to the Postgres database
func ConnAttempts(attempts int) Option {
	return func(c *Postgres) {
		c.connAttempts = attempts
	}
}

// ConnTimeout sets the connection timeout duration for the Postgres database
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connTimeout = timeout
	}
}
