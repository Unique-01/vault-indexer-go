package indexer

import (
	"context"
	"fmt"
	"math/big"
	"math/rand/v2"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
			app.logger.Info("Fetching logs", "fromBlock", job.FromBlock, "toBlock", job.ToBlock)
			logs, err := app.fetchLogsWithRetry(ctx, job)
			if err != nil {
				return fmt.Errorf("Fetch Worker failed: range %d -> %d: %w", job.FromBlock, job.ToBlock, err)
			}

			select {
			case parseChan <- ParseJob{
				RangeJob: job,
				Logs:     logs,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (app *App) fetchLogsWithRetry(ctx context.Context, job RangeJob) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(job.FromBlock),
		ToBlock:   new(big.Int).SetUint64(job.ToBlock),
		Addresses: []common.Address{app.vaultAddress},
	}
	var lastErr error

	for attempt := uint64(0); attempt <= MaxJobRetries; attempt++ {
		logs, err := app.blockchain.FilterLogs(ctx, query)
		if err == nil {
			return logs, nil
		}
		lastErr = err
		if attempt < MaxJobRetries {
			jitter := time.Duration(rand.IntN(500)) * time.Millisecond
			delay := time.Second * time.Duration(1<<attempt)

			timer := time.NewTimer(delay + jitter)

			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-timer.C:
			}
		}

	}

	return nil, fmt.Errorf("max tries exceeded error: %w", lastErr)
}
