package config

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type Config struct {
	RpcUrl       string
	VaultAddress common.Address
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading config: %w", err)
	}
	cfg := &Config{
		RpcUrl:       os.Getenv("RPC_URL"),
		VaultAddress: common.HexToAddress(os.Getenv("VAULT_ADDRESS")),
	}

	return cfg, nil
}
