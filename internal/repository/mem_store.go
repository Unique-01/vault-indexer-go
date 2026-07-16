package repository

import (
	"context"
	"sync"

	"github.com/unique-01/vault-indexer-go/internal/indexer"
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

func (store *MemStore) SaveVaultEvents(ctx context.Context, events []indexer.VaultEvent) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.events = append(store.events, events...)

	return nil

}
