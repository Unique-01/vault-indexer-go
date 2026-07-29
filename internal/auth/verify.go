package auth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func (service *Service) VerifySignature(ctx context.Context, wallet common.Address, signature string) (common.Address, error) {
	nonce, found, err := service.nonceStore.ConsumeNonce(ctx, wallet)
	if err != nil {
		return common.Address{}, fmt.Errorf("consume nonce: %w", err)
	}
	if !found {
		return common.Address{}, ErrInvalidNonce
	}

	message, err := service.buildMessage(wallet, nonce)
	if err != nil {
		return common.Address{}, fmt.Errorf("rebuild message: %w", err)
	}
	if _, err = message.Verify(signature, &service.domain, nil, nil); err != nil {
		return common.Address{}, ErrInvalidSignature
	}
	return wallet, nil
}
