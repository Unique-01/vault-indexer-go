package indexer

import (
	"context"
	"fmt"
	"math/big"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/errgroup"
)

const MaxJobRetries uint64 = 3

func (app *App) fetchWorker(ctx context.Context, rangeChan <-chan RangeJob, parseChan chan<- ParseJob) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job, ok := <-rangeChan:
			if !ok {
				return nil
			}
			// app.logger.Info("Fetching logs", "fromBlock", job.FromBlock, "toBlock", job.ToBlock)
			logs, err := app.fetchLogsWithRetry(ctx, job)
			if err != nil {
				return fmt.Errorf("Fetch Worker failed: range %d -> %d: %w", job.FromBlock, job.ToBlock, err)
			}
			blockTimeStamps, err := app.buildTimestampMap(ctx, logs)
			if err != nil {
				return fmt.Errorf("Block timestamp: %w", err)
			}
			select {
			case parseChan <- ParseJob{
				RangeJob:       job,
				Logs:           logs,
				BlockTimestamps: blockTimeStamps,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (app *App) buildTimestampMap(ctx context.Context, logs []types.Log) (map[uint64]time.Time, error) {
	blocks := uniqueBlockNumber(logs)
	timestamps := make(map[uint64]time.Time)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	var mu sync.Mutex

	for _, blockNumber := range blocks {
		g.Go(func() error {
			timestamp, err := app.fetchHeaderWithRetry(ctx, blockNumber)
			if err != nil {
				return fmt.Errorf("timestamp for block %d: %w", blockNumber, err)
			}
			mu.Lock()
			timestamps[blockNumber] = timestamp
			mu.Unlock()

			return nil
		})

	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return timestamps, nil
}

func (app *App) fetchLogsWithRetry(ctx context.Context, job RangeJob) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(job.FromBlock),
		ToBlock:   new(big.Int).SetUint64(job.ToBlock),
		Addresses: []common.Address{app.vaultAddress},
	}
	var logs []types.Log
	err := app.withRetry(ctx, func() error {
		result, err := app.blockchain.FilterLogs(ctx, query)
		if err != nil {
			return err
		}
		logs = result
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fetch logs range %d -> %d: %w", job.FromBlock, job.ToBlock, err)
	}
	return logs, nil
}

func (app *App) fetchHeaderWithRetry(ctx context.Context, blockNumber uint64) (time.Time, error) {
	var timestamp time.Time

	err := app.withRetry(ctx, func() error {
		header, err := app.blockchain.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
		if err != nil {
			return err
		}
		timestamp = time.Unix(int64(header.Time), 0)
		return nil
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("fetch header block %d: %w", blockNumber, err)
	}
	return timestamp, nil
}

func (app *App) withRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := uint64(0); attempt <= MaxJobRetries; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < MaxJobRetries {
			jitter := time.Duration(rand.IntN(500)) * time.Millisecond
			delay := time.Second * time.Duration(1<<attempt)

			timer := time.NewTimer(delay + jitter)

			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-timer.C:
			}
		}

	}
	return fmt.Errorf("max tries exceeded error: %w", lastErr)
}
func uniqueBlockNumber(logs []types.Log) []uint64 {
	seen := make(map[uint64]bool)
	var blocks []uint64

	for _, log := range logs {
		if !seen[log.BlockNumber] {
			seen[log.BlockNumber] = true
			blocks = append(blocks, log.BlockNumber)
		}
	}
	return blocks
}
