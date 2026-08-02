// internal/indexer/parse_worker_test.go
package indexer

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestBuildVaultEvent(t *testing.T) {
	wallet := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb9226")
	amount := big.NewInt(1000)
	prevAmount := big.NewInt(500)
	newAmount := big.NewInt(700)
	timestamp := time.Now()
	log := types.Log{TxHash: common.HexToHash("0xabc"), Index: 3, BlockNumber: 42}

	tests := []struct {
		name      string
		eventName string
		merged    map[string]any
		wantErr   bool
		check     func(t *testing.T, e VaultEvent)
	}{
		{
			name:      "UserDeposited sets Amount",
			eventName: "UserDeposited",
			merged:    map[string]any{"user": wallet, "amount": amount},
			check: func(t *testing.T, e VaultEvent) {
				if e.Amount.Cmp(amount) != 0 {
					t.Errorf("Amount = %v, want %v", e.Amount, amount)
				}
				if e.PreviousAmount != nil || e.NewAmount != nil {
					t.Errorf("expected PreviousAmount/NewAmount nil, got %v/%v", e.PreviousAmount, e.NewAmount)
				}
			},
		},
		{
			name:      "UserModifiedPendingWithdrawal sets Previous and New",
			eventName: "UserModifiedPendingWithdrawal",
			merged:    map[string]any{"user": wallet, "previousAmount": prevAmount, "newAmount": newAmount},
			check: func(t *testing.T, e VaultEvent) {
				if e.PreviousAmount.Cmp(prevAmount) != 0 || e.NewAmount.Cmp(newAmount) != 0 {
					t.Errorf("got PreviousAmount=%v NewAmount=%v", e.PreviousAmount, e.NewAmount)
				}
				if e.Amount != nil {
					t.Errorf("expected Amount nil, got %v", e.Amount)
				}
			},
		},
		{
			name:      "UserCancelledPendingWithdrawal sets only wallet",
			eventName: "UserCancelledPendingWithdrawal",
			merged:    map[string]any{"user": wallet},
			check: func(t *testing.T, e VaultEvent) {
				if e.Amount != nil || e.PreviousAmount != nil || e.NewAmount != nil {
					t.Errorf("expected all amounts nil")
				}
			},
		},
		{
			name:      "missing user key errors",
			eventName: "UserDeposited",
			merged:    map[string]any{"amount": amount},
			wantErr:   true,
		},
		{
			name:      "user wrong type errors",
			eventName: "UserDeposited",
			merged:    map[string]any{"user": "not-an-address", "amount": amount},
			wantErr:   true,
		},
		{
			name:      "missing amount errors",
			eventName: "UserDeposited",
			merged:    map[string]any{"user": wallet},
			wantErr:   true,
		},
		{
			name:      "unknown event type errors",
			eventName: "SomeFutureEvent",
			merged:    map[string]any{"user": wallet},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := buildVaultEvent(tt.eventName, tt.merged, log, timestamp)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, event)
			}
		})
	}
}
