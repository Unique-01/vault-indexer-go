package indexer

import (
	"context"
	"fmt"
)

func (app *App) saveWorker(ctx context.Context, orderedSaveChan <-chan SaveJob) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job, ok := <-orderedSaveChan:
			if !ok {
				return nil
			}
			err := app.store.SaveRange(ctx, job.ToBlock, job.Events)
			if err != nil {
				return fmt.Errorf("Save vault event range %d -> %d: %w", job.FromBlock, job.ToBlock, err)
			}
			app.logger.Info("saved range", "fromBlock", job.FromBlock, "toBlock", job.ToBlock, "eventCount", len(job.Events))
		}
	}
}
