// Package dbschema defines the schema version required by the API. SQL and
// migration execution live in db/migrations and are linked only by the migrator.
package dbschema

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

const (
	RequiredVersion    = "202610010001" // RequiredVersion is the latest schema version. Must match the latest migration file in db/migrations
	MigrationTableName = "public.bun_migrations"
)

// Check compares the latest applied version with the API requirement
func Check(ctx context.Context, db *bun.DB) error {
	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass(?) IS NOT NULL", MigrationTableName).Scan(&exists); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if !exists {
		return fmt.Errorf("database has no migration history; the %s table does not exist", MigrationTableName)
	}

	var applied string
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(name), '') FROM "+MigrationTableName).Scan(&applied); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if applied != RequiredVersion {
		return fmt.Errorf("incompatible database schema version: applied %q, required %q; run kryvea-migrate before starting the API", applied, RequiredVersion)
	}
	return nil
}
