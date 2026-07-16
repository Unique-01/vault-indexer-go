package indexer

import (
	"context"
)

type Store interface {
	GetLastIndexedBlock(ctx context.Context) (uint64, error)
	SaveVaultEvents(ctx context.Context, events []VaultEvent) error
}
