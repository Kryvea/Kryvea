package db

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"time"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Driver struct {
	db       *bun.DB
	sqlDB    *sql.DB
	filesDir string
	logger   zerolog.Logger
}

func NewDriver(ctx context.Context, dsn, filesDir, adminUser, adminPass string, levelWriter zerolog.LevelWriter) (*Driver, error) {
	logger := zerolog.New(levelWriter).With().
		Str("source", "db-driver").
		Timestamp().Logger()

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	if v := config.GetPgMaxConns(); v > 0 {
		sqlDB.SetMaxOpenConns(int(v))
	}
	if v := config.GetPgMinConns(); v > 0 {
		sqlDB.SetMaxIdleConns(int(v))
	}
	if v := config.GetPgMaxConnLifetime(); v > 0 {
		sqlDB.SetConnMaxLifetime(v)
	}
	if v := config.GetPgMaxConnIdleTime(); v > 0 {
		sqlDB.SetConnMaxIdleTime(v)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		logger.Error().Err(err).Msg("failed to ping postgres")
		return nil, err
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())
	bunDB.RegisterModel(
		(*dbAssessmentTarget)(nil),
		(*dbUserCustomer)(nil),
		(*dbUserAssessment)(nil),
	)
	logger.Debug().Msg("connected to PostgreSQL")

	d := &Driver{db: bunDB, sqlDB: sqlDB, filesDir: filesDir, logger: logger}

	if err := d.applySchema(ctx); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	if err := d.ensureFilesDir(); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("ensure files dir: %w", err)
	}

	if err := d.bootstrapAdmin(ctx, adminUser, adminPass); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("bootstrap admin: %w", err)
	}

	if removed, err := d.FileReference().GCFiles(ctx); err != nil {
		d.logger.Warn().Err(err).Msg("startup file gc failed")
	} else if removed > 0 {
		d.logger.Info().Int("removed", removed).Msg("startup file gc removed orphans")
	}

	return d, nil
}

func (d *Driver) DB() *bun.DB { return d.db }

func (d *Driver) Close() error {
	if d.db != nil {
		_ = d.db.Close()
	}
	if d.sqlDB != nil {
		return d.sqlDB.Close()
	}
	return nil
}

func (d *Driver) ensureFilesDir() error {
	return os.MkdirAll(d.filesDir, 0o755)
}

func (d *Driver) bootstrapAdmin(ctx context.Context, adminUser, adminPass string) error {
	if adminUser == "" || adminPass == "" {
		d.logger.Warn().Msg("admin credentials empty, skipping admin bootstrap")
		return nil
	}
	users := d.User().(*UserIndex)

	existing, err := users.GetByUsername(ctx, adminUser)
	if err != nil && err != store.ErrNotFound {
		return fmt.Errorf("check admin user: %w", err)
	}
	if existing != nil {
		return nil
	}

	if _, err := users.Insert(ctx, &model.User{
		Username:       adminUser,
		Role:           model.RoleAdmin,
		PasswordExpiry: time.Now(),
	}, adminPass); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	d.logger.Info().Str("username", adminUser).Msg("created admin user")
	return nil
}

func advisoryLockKey(name string) int32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	return int32(h.Sum32()) //nolint:gosec // hashing for lock keying, not crypto
}

func (d *Driver) Assessment() store.AssessmentStore       { return &AssessmentIndex{driver: d} }
func (d *Driver) Category() store.CategoryStore           { return &CategoryIndex{driver: d} }
func (d *Driver) Customer() store.CustomerStore           { return &CustomerIndex{driver: d} }
func (d *Driver) FileReference() store.FileReferenceStore { return &FileReferenceIndex{driver: d} }
func (d *Driver) Poc() store.PocStore                     { return &PocIndex{driver: d} }
func (d *Driver) Setting() store.SettingStore             { return &SettingIndex{driver: d} }
func (d *Driver) Target() store.TargetStore               { return &TargetIndex{driver: d} }
func (d *Driver) Template() store.TemplateStore           { return &TemplateIndex{driver: d} }
func (d *Driver) User() store.UserStore                   { return &UserIndex{driver: d} }
func (d *Driver) Vulnerability() store.VulnerabilityStore { return &VulnerabilityIndex{driver: d} }
