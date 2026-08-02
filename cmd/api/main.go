package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/unique-01/vault-indexer-go/internal/api"
	"github.com/unique-01/vault-indexer-go/internal/auth"
	"github.com/unique-01/vault-indexer-go/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, logger, vaultStore, err := bootstrap.Init(ctx)
	if err != nil {
		logger.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}

	authService := auth.NewService(vaultStore, cfg.JwtSecret, cfg.TokenExpiry, cfg.SiweDomain, cfg.SiweURI, cfg.ChainID)
	apiServer := api.New(logger, cfg.APIAddr, vaultStore, authService)

	logger.Info("starting api")
	if err := apiServer.Run(ctx); err != nil {
		logger.Error("shutting down", "error", err)
		os.Exit(1)
	}
}
