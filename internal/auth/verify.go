package auth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spruceid/siwe-go"
)

func (service *Service) VerifySignature(ctx context.Context, wallet common.Address, signature string) (common.Address, error) {
	text, found, err := service.store.ConsumeChallenge(ctx, wallet)
	if err != nil {
		return common.Address{}, fmt.Errorf("consume nonce: %w", err)
	}
	if !found {
		return common.Address{}, ErrInvalidNonce
	}

	message, err := siwe.ParseMessage(text)
	if err != nil {
		return common.Address{}, fmt.Errorf("parse stored message: %w", err)
	}

	if _, err = message.Verify(signature, &service.domain, nil, nil); err != nil {
		return common.Address{}, ErrInvalidSignature
	}
	return wallet, nil
}
