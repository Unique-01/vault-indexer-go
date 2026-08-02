package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/unique-01/vault-indexer-go/internal/bootstrap"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, logger, vaultStore, err := bootstrap.Init(ctx)
	if err != nil {
		logger.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}

	ethClient, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		logger.Error("error connecting to evm node", "error", err)
		os.Exit(1)
	}

	indexerApp, err := indexer.New(cfg, logger, ethClient, vaultStore, indexer.SleepFunc)
	if err != nil {
		logger.Error("app initialization error", "error", err)
		os.Exit(1)
	}

	logger.Info("starting indexer")
	if err := indexerApp.Run(ctx); err != nil {
		logger.Error("shutting down", "error", err)
		os.Exit(1)
	}
}
