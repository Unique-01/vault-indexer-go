package auth

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v5"
)

type sessionClaims struct {
	WalletAddress string `json:"walletAddress"`
	jwt.RegisteredClaims
}

func (service *Service) IssueSession(wallet common.Address) (string, error) {
	now := time.Now()

	claims := sessionClaims{
		WalletAddress: wallet.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(service.tokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(service.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (service *Service) VerifySession(tokenString string) (common.Address, error) {
	claims := &sessionClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return service.jwtSecret, nil
	})
	if err != nil {
		return common.Address{}, fmt.Errorf("%w:%v", ErrInvalidSession, err)
	}
	if !token.Valid {
		return common.Address{}, ErrInvalidSession
	}

	return common.HexToAddress(claims.WalletAddress), nil
}
