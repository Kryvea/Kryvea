package api

import (
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/rs/zerolog"
)

type Driver struct {
	mongo  *mongo.Driver
	logger *zerolog.Logger
}

func NewDriver(mongo *mongo.Driver, levelWriter *zerolog.LevelWriter) *Driver {
	logger := zerolog.New(*levelWriter).With().
		Str("source", "api-driver").
		Timestamp().Logger()

	return &Driver{
		mongo:  mongo,
		logger: &logger,
	}
}
