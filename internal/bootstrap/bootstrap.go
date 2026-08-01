package bootstrap

import (
	"context"
	"fmt"

	"github.com/unique-01/vault-indexer-go/internal/config"
	"github.com/unique-01/vault-indexer-go/internal/repository"
)

func Init(ctx context.Context) (*config.Config, *repository.PostgresStore, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	store, err := repository.NewPostgresStore(ctx, cfg.DatabaseUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres store: %w", err)
	}

	return cfg, store, nil
}
