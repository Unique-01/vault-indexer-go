// internal/repository/postgres_challenges.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func (store *PostgresStore) SaveChallenge(ctx context.Context, wallet common.Address, message string) error {
	const query = `
		INSERT INTO auth_challenges (wallet_address, message, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (wallet_address) DO UPDATE
		SET message = EXCLUDED.message, created_at = EXCLUDED.created_at
	`

	_, err := store.db.ExecContext(ctx, query, wallet.Hex(), message)
	if err != nil {
		return fmt.Errorf("save challenge: %w", err)
	}

	return nil
}

func (store *PostgresStore) ConsumeChallenge(ctx context.Context, wallet common.Address) (string, bool, error) {
	const query = `
		DELETE FROM auth_challenges
		WHERE wallet_address = $1
		RETURNING message
	`

	var message string
	err := store.db.QueryRowContext(ctx, query, wallet.Hex()).Scan(&message)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("consume challenge: %w", err)
	}

	return message, true, nil
}