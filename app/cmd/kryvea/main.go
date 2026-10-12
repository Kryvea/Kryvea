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
	cfg := config.Load()

	if err := i18n.InitI18n(cfg.LocalesPath); err != nil {
		return fmt.Errorf("init i18n: %w", err)
	}

	levelWriter := log.NewLevelWriter(cfg.Log)

	driver, err := db.NewDriver(context.Background(), cfg.DB, cfg.Admin, levelWriter)
	if err != nil {
		return fmt.Errorf("init bun driver: %w", err)
	}

	engine.NewEngine(
		cfg.Addr,
		cfg.RootPath,
		cfg.BodyLimitMB,
		driver,
		levelWriter,
	).Serve()
	return nil
}
