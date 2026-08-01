package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/api"
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

func (store *PostgresStore) ListEvents(ctx context.Context, filter api.EventFilter) (api.EventPage, error) {
	limit := filter.Limit
	if limit <= 0 || limit > maxEventLimit {
		limit = defaultEventLimit
	}

	afterBlock, afterLogIndex := int64(-1), int64(-1)
	if filter.Cursor != nil {
		afterBlock = int64(filter.Cursor.BlockNumber)
		afterLogIndex = int64(filter.Cursor.LogIndex)
	}

	var eventType any
	if filter.EventType != nil {
		eventType = string(*filter.EventType)
	}

	query := `
		SELECT wallet_address, tx_hash, log_index, block_number, event_timestamp, event_type, amount, previous_amount, new_amount
		FROM vault_events
		WHERE wallet_address = $1
		  AND ($2::text IS NULL OR event_type = $2)
		  AND (block_number, log_index) > ($3, $4)
		ORDER BY block_number, log_index
		LIMIT $5
	`

	rows, err := store.db.QueryContext(ctx, query,
		strings.ToLower(filter.WalletAddress.Hex()),
		eventType,
		afterBlock,
		afterLogIndex,
		limit+1,
	)
	if err != nil {
		return api.EventPage{}, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []indexer.VaultEvent
	for rows.Next() {
		var e indexer.VaultEvent
		var walletRaw, txHashRaw, eventTypeRaw string
		var amount, prevAmount, newAmount sql.NullString

		if err := rows.Scan(&walletRaw, &txHashRaw, &e.LogIndex, &e.BlockNumber, &e.TimeStamp, &eventTypeRaw, &amount, &prevAmount, &newAmount); err != nil {
			return api.EventPage{}, fmt.Errorf("scan event: %w", err)
		}

		e.WalletAddress = common.HexToAddress(walletRaw)
		e.TxHash = common.HexToHash(txHashRaw)
		e.EventType = indexer.EventType(eventTypeRaw)
		e.Amount = nullStringToBigInt(amount)
		e.PreviousAmount = nullStringToBigInt(prevAmount)
		e.NewAmount = nullStringToBigInt(newAmount)

		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return api.EventPage{}, fmt.Errorf("rows iteration: %w", err)
	}

	var page api.EventPage
	if len(events) > limit {
		page.Events = events[:limit]
		last := page.Events[len(page.Events)-1]
		page.NextCursor = &api.Cursor{BlockNumber: last.BlockNumber, LogIndex: last.LogIndex}
	} else {
		page.Events = events
	}

	return page, nil
}

func nullStringToBigInt(ns sql.NullString) *big.Int {
	if !ns.Valid {
		return nil
	}
	n := new(big.Int)
	n.SetString(ns.String, 10)
	return n
}
