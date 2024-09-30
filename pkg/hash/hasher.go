package hash

// Hasher interface for password hashing logic
type Hasher interface {
	HashPassword(password string) string
}
