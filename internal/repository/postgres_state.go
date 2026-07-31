package repository

import (
	"context"
	"fmt"

	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

func (store *PostgresStore) GetLastIndexedBlock(ctx context.Context) (uint64, error) {
	var lastIndexed uint64

	row := store.db.QueryRowContext(ctx, `SELECT last_indexed_block FROM indexer_state WHERE id=1`)

	if err := row.Scan(&lastIndexed); err != nil {
		return 0, fmt.Errorf("scan lst indexed block: %w", err)
	}

	return lastIndexed, nil
}

func (store *PostgresStore) SaveRange(ctx context.Context, toBlock uint64, events []indexer.VaultEvent) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := insertEvents(ctx, tx, events); err != nil {
		return fmt.Errorf("insert events: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE indexer_state SET last_indexed_block = $1 WHERE id = 1`, toBlock); err != nil {
		return fmt.Errorf("update last indexed block: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
