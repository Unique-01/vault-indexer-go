package indexer

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type VaultEvent struct {
	WalletAddress common.Address
	TxHash        common.Hash
	LogIndex      uint
	BlockNumber   uint64
	TimeStamp     time.Time
	EventType     string
	Amount        *big.Int
}

type RangeJob struct {
	FromBlock uint64
	ToBlock   uint64
}

type ParseJob struct {
	RangeJob
	Logs []types.Log
}

type SaveJob struct {
	RangeJob
	Events []VaultEvent
}
