package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password with bcrypt
func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// ComparePassword reports whether the password matches the bcrypt hash
func ComparePassword(password string, hash []byte) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}
