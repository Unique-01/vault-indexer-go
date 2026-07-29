package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/unique-01/vault-indexer-go/internal/api"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

const (
	defaultEventLimit = 20
	maxEventLimit     = 100
)

type MemStore struct {
	mu        sync.Mutex
	lastBlock uint64
	events    []indexer.VaultEvent
}

func NewMemStore() *MemStore {
	return &MemStore{}
}

func (store *MemStore) GetLastIndexedBlock(ctx context.Context) (uint64, error) {
	return store.lastBlock, nil
}

func (store *MemStore) SaveRange(ctx context.Context, toBlock uint64, events []indexer.VaultEvent) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.events = append(store.events, events...)
	store.lastBlock = toBlock

	return nil

}

func (store *MemStore) ListEvents(ctx context.Context, filter api.EventFilter) (api.EventPage, error) {
    store.mu.Lock()
    defer store.mu.Unlock()

    limit := clampLimit(filter.Limit)

    var matched []indexer.VaultEvent
    for _, e := range store.events {
        if matchesFilter(e, filter) {
            matched = append(matched, e)
        }
    }

    sort.Slice(matched, func(i, j int) bool {
        return beforeInSequence(matched[i], matched[j])
    })

    return buildPage(matched, limit), nil
}

func clampLimit(limit int) int {
    if limit <= 0 || limit > maxEventLimit {
        return defaultEventLimit
    }
    return limit
}

func matchesFilter(e indexer.VaultEvent, filter api.EventFilter) bool {
    if e.WalletAddress != filter.WalletAddress {
        return false
    }
    if filter.EventType != nil && e.EventType != *filter.EventType {
        return false
    }
    if filter.Cursor != nil && !isAfterCursor(e, *filter.Cursor) {
        return false
    }
    return true
}

func isAfterCursor(e indexer.VaultEvent, c api.Cursor) bool {
    if e.BlockNumber != c.BlockNumber {
        return e.BlockNumber > c.BlockNumber
    }
    return e.LogIndex > c.LogIndex
}

func beforeInSequence(a, b indexer.VaultEvent) bool {
    if a.BlockNumber != b.BlockNumber {
        return a.BlockNumber < b.BlockNumber
    }
    return a.LogIndex < b.LogIndex
}

func buildPage(matched []indexer.VaultEvent, limit int) api.EventPage {
    if len(matched) <= limit {
        return api.EventPage{Events: matched}
    }

    events := matched[:limit]
    last := events[len(events)-1]
    return api.EventPage{
        Events:     events,
        NextCursor: &api.Cursor{BlockNumber: last.BlockNumber, LogIndex: last.LogIndex},
    }
}