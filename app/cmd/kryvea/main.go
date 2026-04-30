package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/Kryvea/Kryvea/internal/db"
	"github.com/Kryvea/Kryvea/internal/engine"
	"github.com/Kryvea/Kryvea/internal/i18n"
	"github.com/Kryvea/Kryvea/internal/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "kryvea: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := i18n.InitI18n(config.GetLocalesPath()); err != nil {
		return fmt.Errorf("init i18n: %w", err)
	}

	levelWriter := log.NewLevelWriter(
		config.GetLogDirectory(),
		config.GetLogMaxSizeMB(),
		config.GetLogMaxBackups(),
		config.GetLogMaxAgeDays(),
		config.GetLogCompress(),
	)

	driver, err := db.NewDriver(
		context.Background(),
		config.GetPgDSN(),
		config.GetFilesDir(),
		config.GetAdminUser(),
		config.GetAdminPass(),
		levelWriter,
	)
	if err != nil {
		return fmt.Errorf("init bun driver: %w", err)
	}

	engine.NewEngine(
		config.GetListeningAddr(),
		config.GetRootPath(),
		config.GetBodyLimitMB(),
		driver,
		levelWriter,
	).Serve()
	return nil
}
