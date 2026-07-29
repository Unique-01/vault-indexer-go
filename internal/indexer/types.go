package indexer

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EventType string

const (
	EventUserDeposited                  EventType = "UserDeposited"
	EventUserRequestedWithdrawal        EventType = "UserRequestedWithdrawal"
	EventUserWithdrawn                  EventType = "UserWithdrawn"
	EventUserModifiedPendingWithdrawal  EventType = "UserModifiedPendingWithdrawal"
	EventUserCancelledPendingWithdrawal EventType = "UserCancelledPendingWithdrawal"
)

func (e EventType) Valid() bool {
	switch e {
	case EventUserDeposited, EventUserRequestedWithdrawal, EventUserWithdrawn,
		EventUserModifiedPendingWithdrawal, EventUserCancelledPendingWithdrawal:
		return true
	default:
		return false
	}
}

type VaultEvent struct {
	WalletAddress common.Address
	TxHash        common.Hash
	LogIndex      uint
	BlockNumber   uint64
	TimeStamp     time.Time
	EventType     EventType

	Amount         *big.Int
	PreviousAmount *big.Int
	NewAmount      *big.Int
}

type RangeJob struct {
	FromBlock uint64
	ToBlock   uint64
}

type ParseJob struct {
	RangeJob
	Logs            []types.Log
	BlockTimestamps map[uint64]time.Time
}

type SaveJob struct {
	RangeJob
	Events []VaultEvent
}
