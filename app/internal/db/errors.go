package db

import (
	"database/sql"
	"errors"

	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/uptrace/bun/driver/pgdriver"
)

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return store.ErrNotFound
	}
	if pgErr, ok := err.(pgdriver.Error); ok {
		switch pgErr.Field('C') {
		case "23505":
			return store.ErrDuplicateKey
		case "23503":
			return store.ErrFKViolation
		case "23502":
			return store.ErrNotNullViolation
		case "40P01":
			return store.ErrDeadlock
		case "55P03":
			return store.ErrLockNotAvailable
		}
	}
	return err
}
