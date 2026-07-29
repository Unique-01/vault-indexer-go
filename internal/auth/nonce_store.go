package auth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type NonceStore interface {
	SaveNonce(ctx context.Context, wallet common.Address, nonce string) error
	ConsumeNonce(ctx context.Context, wallet common.Address) (nonce string, found bool, err error)
}
