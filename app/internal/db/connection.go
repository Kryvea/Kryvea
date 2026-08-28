package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func ConnectDB(ctx context.Context, cfg config.DB, levelWriter zerolog.LevelWriter) (*bun.DB, error) {
	logger := zerolog.New(levelWriter).With().
		Str("source", "db-connection").
		Timestamp().Logger()

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DSN)))

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())
	bunDB.RegisterModel(
		(*dbAssessmentTarget)(nil),
		(*dbUserCustomer)(nil),
		(*dbUserAssessment)(nil),
	)

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := bunDB.PingContext(pingCtx); err != nil {
		_ = bunDB.Close()
		logger.Error().Err(err).Msg("failed to ping postgres")
		return nil, err
	}

	logger.Debug().Msg("connected to PostgreSQL")

	return bunDB, nil
}
