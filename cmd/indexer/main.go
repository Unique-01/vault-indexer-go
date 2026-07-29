package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/unique-01/vault-indexer-go/internal/api"
	"github.com/unique-01/vault-indexer-go/internal/config"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
	"github.com/unique-01/vault-indexer-go/internal/repository"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := slog.Default()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "error", err)
		os.Exit(1)
	}

	ethClient, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		logger.Error("Error connecting to evm node", "error", err)
		os.Exit(1)
	}
	vaultStore := repository.NewMemStore()

	indexerApp, err := indexer.New(cfg, logger, ethClient, vaultStore)
	if err != nil {
		logger.Error("App initialization error", "error", err)
		os.Exit(1)
	}

	apiServer := api.New(logger, cfg.APIAddr, vaultStore)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return indexerApp.Run(ctx)
	})
	g.Go(func() error {
		return apiServer.Run(ctx)
	})

	logger.Info("Starting indexer and Api")
	if err := g.Wait(); err != nil {
		logger.Error("Shutting down", "error", err)
		os.Exit(1)
	}
}
