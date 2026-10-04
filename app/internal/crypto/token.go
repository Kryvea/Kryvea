package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

const (
	tokenLength int = 32
)

// Token represents a cryptographically secure random token
type Token []byte

// String returns the token as a base64url encoded string
func (t Token) String() string {
	return base64.RawURLEncoding.EncodeToString(t)
}

// NewToken generates a cryptographically secure random token
func NewToken() Token {
	bytes := make([]byte, tokenLength)
	rand.Read(bytes)
	return Token(bytes)
}

// ParseToken parses a base64url encoded string and returns a Token if valid
func ParseToken(s string) (Token, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(bytes) != tokenLength {
		return nil, errors.New("invalid token length")
	}

	return Token(bytes), nil
}
