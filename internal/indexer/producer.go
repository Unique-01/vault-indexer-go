package indexer

import (
	"context"
	"fmt"
)

func (app *App) rangeProducer(ctx context.Context, rangeChan chan<- RangeJob, batchSize uint64) error {
	for {

		lastIndexed, err := app.store.GetLastIndexedBlock(ctx)
		if err != nil {
			return fmt.Errorf("get last indexed block: %w", err)
		}

		latestBlock, err := app.blockchain.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("get latest block: %w", err)
		}

		for fromBlock := lastIndexed + 1; fromBlock <= latestBlock; fromBlock += batchSize {
			toBlock := min(fromBlock+batchSize-1, latestBlock)

			select {
			case rangeChan <- RangeJob{
				FromBlock: fromBlock,
				ToBlock:   toBlock,
			}:
			case <-ctx.Done():
				return ctx.Err()

			}

		}
		// Implement waiting interval

	}
}
