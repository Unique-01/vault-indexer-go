package api

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

type EventFilter struct {
	WalletAddress common.Address
	EventType     *indexer.EventType
	Cursor        *Cursor
	Limit         int
}

type EventPage struct {
	Events     []indexer.VaultEvent
	NextCursor *Cursor
}

type EventReader interface {
	ListEvents(ctx context.Context, filter EventFilter) (EventPage, error)
}
