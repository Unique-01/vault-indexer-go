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
	RPCUrl       string
	VaultAddress common.Address
	BatchSize    uint64
	PollInterval time.Duration
	APIAddr      string
	JwtSecret    []byte
	TokenExpiry  time.Duration
	SiweDomain   string
	SiweURI      string
	ChainID      int
	DatabaseUrl  string
	Env          string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	pollInterval, err := parseDurationEnv("POLL_INTERVAL")
	if err != nil {
		return nil, err
	}

	batchSize, err := parseUintEnv("BATCH_SIZE")
	if err != nil {
		return nil, err
	}

	tokenExpiry, err := parseDurationEnv("JWT_TOKEN_EXPIRY")
	if err != nil {
		return nil, err
	}

	chainID, err := parseIntEnv("CHAIN_ID")
	if err != nil {
		return nil, err
	}

	vaultAddressRaw := os.Getenv("VAULT_ADDRESS")
	if !common.IsHexAddress(vaultAddressRaw) {
		return nil, fmt.Errorf("invalid or missing VAULT_ADDRESS")
	}

	return &Config{
		RPCUrl:       os.Getenv("RPC_URL"),
		VaultAddress: common.HexToAddress(vaultAddressRaw),
		BatchSize:    batchSize,
		PollInterval: pollInterval,
		APIAddr:      os.Getenv("HTTP_ADDR"),
		JwtSecret:    []byte(os.Getenv("JWT_SECRET")),
		TokenExpiry:  tokenExpiry,
		SiweDomain:   os.Getenv("SIWE_DOMAIN"),
		SiweURI:      os.Getenv("SIWE_URI"),
		ChainID:      chainID,
		DatabaseUrl:  os.Getenv("DATABASE_URL"),
		Env:          os.Getenv("APP_ENV"),
	}, nil
}

func parseDurationEnv(key string) (time.Duration, error) {
	val, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return val, nil
}

func parseUintEnv(key string) (uint64, error) {
	val, err := strconv.ParseUint(os.Getenv(key), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return val, nil
}

func parseIntEnv(key string) (int, error) {
	val, err := strconv.ParseInt(os.Getenv(key), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return int(val), nil
}
