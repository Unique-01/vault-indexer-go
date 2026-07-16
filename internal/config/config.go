package config

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	RpcUrl       string
	VaultAddress common.Address
}

func Load() *Config {
	cfg := &Config{
		RpcUrl:       os.Getenv("RPC_URL"),
		VaultAddress: common.HexToAddress(os.Getenv("VAULT_ADDRESS")),
	}

	return cfg
}
