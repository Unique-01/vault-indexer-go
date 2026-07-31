package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

func insertEvents(ctx context.Context, tx *sql.Tx, events []indexer.VaultEvent) error {
	if len(events) == 0 {
		return nil
	}

	valueStrings := make([]string, len(events))
	args := make([]any, 0, len(events)*9)

	for i, e := range events {
		n := i * 9
		valueStrings[i] = fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9)

		args = append(args,
			strings.ToLower(e.WalletAddress.Hex()),
			strings.ToLower(e.TxHash.Hex()),
			e.LogIndex,
			e.BlockNumber,
			e.TimeStamp,
			string(e.EventType),
			numericOrNil(e.Amount),
			numericOrNil(e.PreviousAmount),
			numericOrNil(e.NewAmount),
		)
	}

	query := `INSERT INTO vault_events
			(wallet_address,tx_hash,log_index,block_number,event_timestamp,event_type,amount,previous_amount,new_amount)
			VALUES` + strings.Join(valueStrings, ",") + `
			ON CONFLICT (tx_hash,log_index) DO NOTHING`

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("bulk inserts %d events: %w", len(events), err)
	}
	return nil
}

func numericOrNil(v *big.Int) any {
	if v == nil {
		return nil
	}
	return v.String()
}
