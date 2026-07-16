package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/unique-01/vault-indexer-go/internal/config"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
	"github.com/unique-01/vault-indexer-go/internal/repository"
)

func main() {
	logger := slog.Default()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}
	cfg := config.Load()

	ethClient, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		logger.Error("Error connecting to evm node", "error", err)
	}
	vaultStore := repository.NewMemStore()
	app := indexer.New(cfg, logger, ethClient, vaultStore)

	logger.Info("Starting indexer application")
	if err := app.Run(ctx); err != nil {
		logger.Error("Indexer shutting down", "error", err)
		os.Exit(1)
	}
}
