package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/unique-01/vault-indexer-go/internal/config"
	"github.com/unique-01/vault-indexer-go/internal/repository"
)

func Init(ctx context.Context) (*config.Config, *slog.Logger, *repository.PostgresStore, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}

	store, err := repository.NewPostgresStore(ctx, cfg.DatabaseUrl)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create postgres store: %w", err)
	}

	var slogHandler slog.Handler
	if cfg.Env == "production" {
		slogHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		slogHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(slogHandler)

	return cfg, logger, store, nil
}
