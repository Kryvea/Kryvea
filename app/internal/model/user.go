package model

import (
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/google/uuid"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	TokenExpireTime         = 9 * time.Hour
	TokenExpireTimePwdReset = 15 * time.Minute
	TokenExtendTime         = 2 * time.Hour
	TokenRefreshThreshold   = 1 * time.Hour
)

var (
	Roles     = []string{RoleAdmin, RoleUser}
	TimeNever = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)

type User struct {
	Model
	DisabledAt     time.Time    `json:"disabled_at,omitempty"`
	Username       string       `json:"username"`
	Password       []byte       `json:"-"`
	PasswordExpiry time.Time    `json:"-"`
	Token          crypto.Token `json:"-"`
	TokenExpiry    time.Time    `json:"-"`
	Role           string       `json:"role"`
	Customers      []Customer   `json:"customers,omitempty"`
	Assessments    []Assessment `json:"assessments,omitempty"`
}

func (u *User) CanAccessCustomer(customer uuid.UUID) bool {
	if u.Role == RoleAdmin {
		return true
	}
	for _, allowedCustomer := range u.Customers {
		if allowedCustomer.ID == customer {
			return true
		}
	}
	return false
}

func IsValidRole(role string) bool {
	if role == "" {
		return false
	}
	for _, r := range Roles {
		if r == role {
			return true
		}
	}
	return false
}
