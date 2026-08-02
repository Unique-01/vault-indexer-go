package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/config"
	"golang.org/x/sync/errgroup"
)

type sleepFunc func(ctx context.Context, d time.Duration) error
type App struct {
	logger       *slog.Logger
	blockchain   BlockchainClient
	store        Store
	vaultAddress common.Address
	contractABI  abi.ABI
	pollInterval time.Duration
	batchSize    uint64
	sleep        sleepFunc
}

func New(cfg *config.Config, logger *slog.Logger, blockchain BlockchainClient, store Store, sleep sleepFunc) (*App, error) {
	contractABI, err := loadVaultABI()
	if err != nil {
		return nil, fmt.Errorf("load vault Abi: %w", err)
	}
	return &App{
		logger:       logger,
		blockchain:   blockchain,
		store:        store,
		vaultAddress: cfg.VaultAddress,
		contractABI:  contractABI,
		pollInterval: cfg.PollInterval,
		batchSize:    cfg.BatchSize,
		sleep:        sleep,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	rootG, ctx := errgroup.WithContext(ctx)

	rangeChan := make(chan RangeJob, 20)
	parseChan := make(chan ParseJob, 20)
	saveChan := make(chan SaveJob, 20)
	orderedSaveChan := make(chan SaveJob, 20)

	rootG.Go(func() error {
		defer close(rangeChan)
		return app.rangeProducer(ctx, rangeChan, app.batchSize, app.pollInterval)
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

	rootG.Go(func() error {
		defer close(saveChan)
		parseG, parseCtx := errgroup.WithContext(ctx)
		for range 3 {
			parseG.Go(func() error {
				return app.parseWorker(parseCtx, parseChan, saveChan)
			})
		}
		return parseG.Wait()
	})

	rootG.Go(func() error {
		defer close(orderedSaveChan)
		return app.sequencer(ctx, saveChan, orderedSaveChan)
	})

	rootG.Go(func() error {
		return app.saveWorker(ctx, orderedSaveChan)
	})

	return rootG.Wait()

}
