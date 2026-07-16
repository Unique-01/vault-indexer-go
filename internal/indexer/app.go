package indexer

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/config"
	"golang.org/x/sync/errgroup"
)

type App struct {
	logger       *slog.Logger
	blockchain   BlockchainClient
	store        Store
	vaultAddress common.Address
}

func New(cfg *config.Config, logger *slog.Logger, blockchain BlockchainClient, store Store) *App {

	return &App{
		logger:       logger,
		blockchain:   blockchain,
		store:        store,
		vaultAddress: cfg.VaultAddress,
	}
}

func (app *App) Run(ctx context.Context) error {
	rootG, ctx := errgroup.WithContext(ctx)

	const batchSize uint64 = 10
	rangeChan := make(chan RangeJob, 20)
	parseChan := make(chan ParseJob, 20)

	rootG.Go(func() error {
		defer close(rangeChan)
		return app.rangeProducer(ctx, rangeChan, batchSize)
	})

	rootG.Go(func() error {
		defer close(parseChan)
		fetchG, fetchCtx := errgroup.WithContext(ctx)
		for range 5 {
			fetchG.Go(func() error {
				return app.fetchWorker(fetchCtx, rangeChan, parseChan)
			})
		}
		return fetchG.Wait()
	})

	return rootG.Wait()

}
