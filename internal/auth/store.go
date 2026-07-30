package auth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type Store interface {
	SaveChallenge(ctx context.Context, wallet common.Address, message string) error
	ConsumeChallenge(ctx context.Context, wallet common.Address) (string, bool, error)
}
