package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spruceid/siwe-go"
)

type Service struct {
	store       Store
	jwtSecret   []byte
	tokenExpiry time.Duration
	domain      string
	uri         string
	chainId     int
}

func NewService(Store Store, jwtSecret []byte, tokenExpiry time.Duration, domain, uri string, chainId int) *Service {
	return &Service{
		store:       Store,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
		domain:      domain,
		uri:         uri,
		chainId:     chainId,
	}
}

func (service *Service) IssueChallenge(ctx context.Context, wallet common.Address) (string, error) {
	nonce := siwe.GenerateNonce()

	message, err := service.buildMessage(wallet, nonce)
	if err != nil {
		return "", fmt.Errorf("build message: %w", err)
	}
	challengeText := message.String()
	if err := service.store.SaveChallenge(ctx, wallet, challengeText); err != nil {
		return "", fmt.Errorf("save challenge: %w", err)
	}

	return challengeText, nil
}

func (service *Service) buildMessage(wallet common.Address, nonce string) (*siwe.Message, error) {
	return siwe.InitMessage(
		service.domain,
		wallet.Hex(),
		service.uri,
		nonce,
		map[string]any{
			"chainId":   service.chainId,
			"statement": "Sign in to Vault Indexer",
		},
	)
}
