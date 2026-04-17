package util

import (
	"testing"
)

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"Valid password", "Password1@", true},
		{"No uppercase letter", "password1@", false},
		{"No lowercase letter", "PASSWORD1@", false},
		{"No digit", "Password@", false},
		{"No special character", "Password1", false},
		{"Too short", "P1@", false},
		{"Empty password", "", false},
		{"With space", "Password 1", true},
		{"With punctuation", "Password.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPassword(tt.password); got != tt.want {
				t.Errorf("IsValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
