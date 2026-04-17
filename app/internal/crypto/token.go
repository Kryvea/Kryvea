package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

const (
	tokenLength int = 32
)

var (
	TokenNil Token
)

// Token represents a cryptographically secure random token
type Token []byte

// String returns the token as a base64url encoded string
func (t Token) String() string {
	return base64.RawURLEncoding.EncodeToString(t)
}

// Bytes returns the raw bytes of the token
func (t Token) Bytes() []byte {
	return []byte(t)
}

// New generates a cryptographically secure random token
func NewToken() Token {
	bytes := make([]byte, tokenLength)
	rand.Read(bytes)
	return Token(bytes)
}

// Parse parses a base64url encoded string and returns a Token if valid
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
