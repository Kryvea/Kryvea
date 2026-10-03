package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/Kryvea/Kryvea/internal/db"
	"github.com/Kryvea/Kryvea/internal/db/migrations"
	"github.com/rs/zerolog"
)

const migrationTimeout = 5 * time.Minute

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "kryvea-migrate: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer) error {
	if len(args) != 1 || (args[0] != "up" && args[0] != "status") {
		return fmt.Errorf("usage: kryvea-migrate <up|status>")
	}

	ctx, cancel := context.WithTimeout(ctx, migrationTimeout)
	defer cancel()

	cfg := config.Load().DB
	if cfg.MaxOpenConns == 1 {
		cfg.MaxOpenConns = 2
	}

	writer := zerolog.MultiLevelWriter(output)
	logger := zerolog.New(writer).With().Timestamp().Str("source", "migrate").Logger()

	connection, err := db.ConnectDB(ctx, cfg, writer)
	if err != nil {
		return err
	}
	defer connection.Close()

	if args[0] == "status" {
		return migrations.Status(ctx, connection, logger)
	}

	return migrations.Run(ctx, connection, logger)
}
