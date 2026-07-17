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
	if pgErr, ok := err.(pgdriver.Error); ok && pgErr.Field('C') == "23505" {
		return store.ErrDuplicateKey
	}
	return err
}
