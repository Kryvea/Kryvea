// Package migrations is used by the standalone migration binary, not API startup.
package migrations

import (
	"cmp"
	"context"
	"database/sql/driver"
	"embed"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Kryvea/Kryvea/internal/dbschema"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

const (
	migrationLocksTableName string = "public.bun_migration_locks"
	migrationLockNamespace  int32  = 0x4C4F434B // LOCK
	migrationLockID         int32  = 0x4D494752 // MIGR
)

//go:embed *.sql
var files embed.FS

func loadMigrations() (*migrate.Migrations, error) {
	m := migrate.NewMigrations()

	if err := m.Discover(files); err != nil {
		return nil, fmt.Errorf("discover migrations: %w", err)
	}

	sorted := m.Sorted()
	if len(sorted) == 0 {
		return nil, errors.New("no migrations found")
	}

	latest := sorted[len(sorted)-1].Name
	if latest != dbschema.RequiredVersion {
		return nil, fmt.Errorf("latest migration %q differs from required API schema version %q", latest, dbschema.RequiredVersion)
	}

	return m, nil
}

// Run applies pending migrations. The caller must allow at least two open
// connections: one dedicated to the session lock and one for migration work.
func Run(ctx context.Context, db *bun.DB, logger zerolog.Logger) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	return run(ctx, db, migrations, logger)
}

func run(ctx context.Context, db *bun.DB, migrations *migrate.Migrations, logger zerolog.Logger) (err error) {
	if db.Stats().MaxOpenConnections == 1 {
		return errors.New("migration execution requires at least two database connections")
	}

	unlock, err := lock(ctx, db)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, unlock()) }()

	m := migrate.NewMigrator(db, migrations,
		migrate.WithTableName(dbschema.MigrationTableName),
		migrate.WithLocksTableName(migrationLocksTableName),
		migrate.WithMarkAppliedOnSuccess(true),
		migrate.BeforeMigration(func(_ context.Context, _ bun.IConn, migration *migrate.Migration) error {
			logger.Info().Str("migration", migration.String()).Msg("applying migration")
			return nil
		}),
	)

	if err := m.Init(ctx); err != nil {
		return fmt.Errorf("initialize migration history: %w", err)
	}

	if err := validatePreviousMigrations(ctx, m, migrations); err != nil {
		return err
	}

	migrationGroup, err := m.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	logger.Info().Int("applied", len(migrationGroup.Migrations)).Msg("migrations complete")

	return nil
}

// Status reports pending and applied migrations without initializing metadata.
func Status(ctx context.Context, db *bun.DB, logger zerolog.Logger) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass(?) IS NOT NULL", dbschema.MigrationTableName).Scan(&exists); err != nil {
		return err
	}

	migrationsSorted := migrations.Sorted()
	if exists {
		m := migrate.NewMigrator(db, migrations, migrate.WithTableName(dbschema.MigrationTableName))

		if err := validatePreviousMigrations(ctx, m, migrations); err != nil {
			return err
		}

		migrationsSorted, err = m.MigrationsWithStatus(ctx)
		if err != nil {
			return err
		}
	}

	for _, migration := range migrationsSorted {
		state := "pending"
		if migration.IsApplied() {
			state = "applied"
		}

		logger.Info().Str("migration", migration.String()).Str("status", state).Msg("migration status")
	}

	return nil
}

// validatePreviousMigrations checks that previously applied migrations are consistent with the discovered migrations
func validatePreviousMigrations(ctx context.Context, m *migrate.Migrator, migrations *migrate.Migrations) error {
	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("read migration history: %w", err)
	}

	// sort by migration name in ascending order
	slices.SortFunc(applied, func(a, b migrate.Migration) int { return cmp.Compare(a.Name, b.Name) })
	migrationsSorted := migrations.Sorted()

	if len(applied) > len(migrationsSorted) {
		return fmt.Errorf("migration history has %d entries [%s], but only %d migrations were discovered", len(applied), applied.String(), len(migrationsSorted))
	}

	for i, migration := range applied {
		if migration.Name != migrationsSorted[i].Name {
			return fmt.Errorf("migration history has version %q at position %d, expected %q", migration.Name, i+1, migrationsSorted[i].Name)
		}
	}

	return nil
}

func lock(ctx context.Context, db *bun.DB) (func() error, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("open migration lock connection: %w", err)
	}

	lockCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		var acquired bool

		if err := conn.QueryRowContext(lockCtx, "SELECT pg_try_advisory_lock(?, ?)", migrationLockNamespace, migrationLockID).Scan(&acquired); err != nil {
			// Force closure of the underlying connection.
			// This ends the PostgreSQL session and frees the lock
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			_ = conn.Close()

			return nil, fmt.Errorf("acquire migration lock: %w", err)
		}

		if acquired {
			break
		}

		select {
		case <-lockCtx.Done():
			_ = conn.Close()
			return nil, fmt.Errorf("wait for migration lock: %w", lockCtx.Err())
		case <-ticker.C:
		}
	}

	// release function
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var released bool

		err := conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock(?, ?)", migrationLockNamespace, migrationLockID).Scan(&released)
		if err != nil || !released {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			err = errors.Join(err, errors.New("failed to release migration lock"))
		}

		return errors.Join(err, conn.Close())
	}, nil
}
