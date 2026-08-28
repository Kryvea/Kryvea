package db

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"time"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type Driver struct {
	db       *bun.DB
	filesDir string
	logger   zerolog.Logger
}

func NewDriver(ctx context.Context, cfg config.DB, admin config.Admin, levelWriter zerolog.LevelWriter) (*Driver, error) {
	logger := zerolog.New(levelWriter).With().
		Str("source", "db-driver").
		Timestamp().Logger()

	bunDB, err := ConnectDB(ctx, cfg, levelWriter)
	if err != nil {
		return nil, err
	}

	d := &Driver{db: bunDB, filesDir: cfg.FilesDir, logger: logger}

	if err = d.initializeApplication(ctx, admin); err != nil {
		_ = d.Close()
		return nil, err
	}

	return d, nil
}

func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *Driver) initializeApplication(ctx context.Context, admin config.Admin) error {
	if err := d.applySchema(ctx); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	if err := d.ensureFilesDir(); err != nil {
		return fmt.Errorf("ensure files dir: %w", err)
	}

	if err := d.bootstrapAdmin(ctx, admin.User, admin.Pass); err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}

	if removed, err := d.FileReference().GCFiles(ctx); err != nil {
		d.logger.Warn().Err(err).Msg("startup file gc failed")
	} else if removed > 0 {
		d.logger.Info().Int("removed", removed).Msg("startup file gc removed orphans")
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
	users := &UserIndex{driver: d}

	existing, err := users.GetByUsername(ctx, adminUser)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
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
