package indexer

import (
	"context"
)

type Store interface {
	GetLastIndexedBlock(ctx context.Context) (uint64, error)
	SaveRange(ctx context.Context, toBlock uint64, events []VaultEvent) error
}
