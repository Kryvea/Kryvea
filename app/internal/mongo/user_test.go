package mongo

import (
	"testing"

	"github.com/google/uuid"
)

func TestCanAccessCustomer(t *testing.T) {
	adminUser := &User{
		Role: RoleAdmin,
	}
	regularUser := &User{
		Role: RoleUser,
		Customers: []Customer{
			{Model: Model{ID: uuid.New()}},
			{Model: Model{ID: uuid.New()}},
		},
	}
	existingCustomer := regularUser.Customers[0].ID
	nonExistingCustomer := uuid.New()

	tests := []struct {
		name     string
		user     *User
		customer uuid.UUID
		want     bool
	}{
		{"Admin user", adminUser, uuid.New(), true},
		{"Regular user with access", regularUser, existingCustomer, true},
		{"Regular user without access", regularUser, nonExistingCustomer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.CanAccessCustomer(tt.customer); got != tt.want {
				t.Errorf("CanAccessCustomer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{"Valid role", RoleAdmin, true},
		{"Invalid role", "invalid_role", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.want {
				t.Errorf("IsValidRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
