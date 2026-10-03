package mongo

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrImmutableCategory      error = errors.New("cannot edit immutable category")
	ErrImmutableTarget        error = errors.New("cannot edit immutable target")
	ErrFileSizeTooLarge       error = errors.New("file size is too large")
	ErrTemplateFileIDRequired error = errors.New("template file ID is required")
	ErrUsedByIDRequired       error = errors.New("used by ID is required")
	ErrAdminUserRequired      error = errors.New("at least one admin user is required")
	ErrLocked                 error = errors.New("lock is locked")
)

func IsDuplicateKeyError(err error) bool {
	return mongo.IsDuplicateKeyError(err)
}
