package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"otome-server/internal/platform/cache"
	"otome-server/internal/platform/database"
	"otome-server/internal/platform/logger"
	"otome-server/internal/shared/config"
	httpapi "otome-server/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.Env, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.MySQLDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	rdb, err := cache.Open(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return err
	}
	defer rdb.Close()

	srv := httpapi.New(cfg, log, db, rdb)
	return srv.Run(ctx)
}
