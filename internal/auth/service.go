package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spruceid/siwe-go"
)

type Service struct {
	nonceStore  NonceStore
	jwtSecret   []byte
	tokenExpiry time.Duration
	domain      string
	uri         string
	chainId     int
}

func NewService(nonceStore NonceStore, jwtSecret []byte, tokenExpiry time.Duration, domain, uri string, chainId int) *Service {
	return &Service{
		nonceStore:  nonceStore,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
		domain:      domain,
		uri:         uri,
		chainId:     chainId,
	}
}

func (service *Service) IssueChallenge(ctx context.Context, wallet common.Address) (string, error) {
	nonce := siwe.GenerateNonce()
	if err := service.nonceStore.SaveNonce(ctx, wallet, nonce); err != nil {
		return "", fmt.Errorf("save nonce: %w", err)
	}

	message, err := service.buildMessage(wallet, nonce)
	if err != nil {
		return "", fmt.Errorf("build message: %w", err)
	}

	return message.String(), nil
}

func (service *Service) buildMessage(wallet common.Address, nonce string) (*siwe.Message, error) {
	return siwe.InitMessage(
		service.domain,
		wallet.Hex(),
		service.uri,
		nonce,
		map[string]interface{}{
			"chainId":   service.chainId,
			"statement": "Sign in to Vault Indexer",
		},
	)
}
