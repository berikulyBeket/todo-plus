package hash

import (
	"crypto/sha1"
	"fmt"
)

// SHA1Hasher provides password hashing functionality using SHA1 with a salt
type SHA1Hasher struct {
	salt string
}

// New creates a new SHA1Hasher with the provided salt
func New(salt string) Hasher {
	return &SHA1Hasher{salt: salt}
}

// HashPassword hashes the provided password using SHA1 and appends the salt
func (h *SHA1Hasher) HashPassword(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(h.salt)))
}
