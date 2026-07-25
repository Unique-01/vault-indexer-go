package indexer

import (
	"context"
	"fmt"
)

func (app *App) sequencer(ctx context.Context, saveChan <-chan SaveJob, orderedSaveChan chan<- SaveJob) error {
	lastIndexed, err := app.store.GetLastIndexedBlock(ctx)
	if err != nil {
		return fmt.Errorf("get last indexed block: %w", err)
	}

	nextExpectedBlock := lastIndexed + 1

	pending := make(map[uint64]SaveJob)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job, ok := <-saveChan:
			if !ok {
				return nil
			}
			pending[job.FromBlock] = job

			for {
				readyJob, found := pending[nextExpectedBlock]
				if !found {
					break
				}
				// app.logger.Info("Sequencing", "fromBlock", readyJob.FromBlock, "toBlock", readyJob.ToBlock)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case orderedSaveChan <- readyJob:
				}

				delete(pending, nextExpectedBlock)
				nextExpectedBlock = readyJob.ToBlock + 1
			}
		}
	}
}
