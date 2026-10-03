package api

import (
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/rs/zerolog"
)

type Driver struct {
	db     store.Store
	logger zerolog.Logger
}

func NewDriver(db store.Store, levelWriter zerolog.LevelWriter) *Driver {
	logger := zerolog.New(levelWriter).With().
		Str("source", "api-driver").
		Timestamp().Logger()

	return &Driver{
		db:     db,
		logger: logger,
	}
}
