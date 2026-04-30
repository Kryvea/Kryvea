package store

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
)

type UserStore interface {
	Insert(ctx context.Context, user *model.User, password string) (uuid.UUID, error)

	Login(ctx context.Context, username, password string) (*model.User, error)
	RefreshUserToken(ctx context.Context, user *model.User) error
	Logout(ctx context.Context, ID uuid.UUID) error

	Get(ctx context.Context, ID uuid.UUID) (*model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
	GetAllUsernames(ctx context.Context) ([]string, error)
	GetByToken(ctx context.Context, token crypto.Token) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)

	Update(ctx context.Context, ID uuid.UUID, user *model.User) error
	UpdateMe(ctx context.Context, userID uuid.UUID, newUser *model.User, password string) error
	UpdateOwnedAssessment(ctx context.Context, userID, assessmentID uuid.UUID, addToOwned bool) error
	AssignCustomer(ctx context.Context, userID, customerID uuid.UUID) error

	Delete(ctx context.Context, ID uuid.UUID) error

	ResetUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	ResetPassword(ctx context.Context, user *model.User, password string) error
	ValidatePassword(ctx context.Context, ID uuid.UUID, currentPassword string) error
}
