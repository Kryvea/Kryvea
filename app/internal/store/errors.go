package store

import "errors"

var (
	ErrImmutableCategory      = errors.New("cannot edit immutable category")
	ErrImmutableTarget        = errors.New("cannot edit immutable target")
	ErrFileSizeTooLarge       = errors.New("file size is too large")
	ErrTemplateFileIDRequired = errors.New("template file ID is required")
	ErrAdminUserRequired      = errors.New("at least one admin user is required")
	ErrLocked                 = errors.New("lock is locked")
	ErrImageTypeNotAllowed    = errors.New("image type not allowed")
	ErrDisabledUser           = errors.New("user is disabled")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidSortField       = errors.New("invalid sort_field")
	ErrNotFound               = errors.New("not found")
	ErrDuplicateKey           = errors.New("duplicate key")
	ErrFKViolation            = errors.New("foreign key violation")
	ErrNotNullViolation       = errors.New("not null violation")
	ErrDeadlock               = errors.New("deadlock detected")
	ErrLockNotAvailable       = errors.New("lock not available")
)
