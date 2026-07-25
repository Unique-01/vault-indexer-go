package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type Config struct {
	RpcUrl       string
	VaultAddress common.Address
	BatchSize    uint64
	PollInterval time.Duration
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading config: %w", err)
	}
	pollInterval, err := time.ParseDuration(os.Getenv("POLL_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("Parse time duration: %w", err)
	}
	batchSize, err := strconv.ParseUint(os.Getenv("BATCH_SIZE"), 10, 0)
	if err != nil {
		return nil, fmt.Errorf("ParseUint: %w", err)
	}
	cfg := &Config{
		RpcUrl:       os.Getenv("RPC_URL"),
		VaultAddress: common.HexToAddress(os.Getenv("VAULT_ADDRESS")),
		BatchSize:    batchSize,
		PollInterval: pollInterval,
	}

	return cfg, nil
}
