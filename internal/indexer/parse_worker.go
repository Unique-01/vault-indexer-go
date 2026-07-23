package indexer

import (
	"context"
	"fmt"
	"maps"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (app *App) parseWorker(ctx context.Context, parseChan <-chan ParseJob, saveChan chan<- SaveJob) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job, ok := <-parseChan:
			if !ok {
				return nil
			}
			var events []VaultEvent
			app.logger.Info("Parsing logs", "fromBlock", job.FromBlock, "toBlock", job.ToBlock)

			for _, log := range job.Logs {
				timestamp, ok := job.BlockTimestamps[log.BlockNumber]
				if !ok {
					app.logger.Error("missing timestamp for block", "block", log.BlockNumber)
					continue
				}
				event, err := app.parseLog(log, timestamp)
				if err != nil {
					return fmt.Errorf("parse log failed: txHash: %s, logIndex: %d, error: %w", log.TxHash, log.Index, err)
				}

				events = append(events, event)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case saveChan <- SaveJob{RangeJob: job.RangeJob, Events: events}:
			}
		}
	}
}

func (app *App) parseLog(log types.Log, timestamp time.Time) (VaultEvent, error) {
	event, err := app.matchEvent(log)
	if err != nil {
		return VaultEvent{}, fmt.Errorf("match event: %w", err)
	}

	data, err := app.unpackEventData(event, log)
	if err != nil {
		return VaultEvent{}, fmt.Errorf("unpack event data: %w", err)
	}

	topics, err := app.unpackEventTopics(event, log)
	if err != nil {
		return VaultEvent{}, fmt.Errorf("unpack event topics: %w", err)
	}

	merged := mergeArgs(data, topics)
	return buildVaultEvent(event.Name, merged, log, timestamp)

}

func (app *App) matchEvent(log types.Log) (*abi.Event, error) {
	if len(log.Topics) == 0 {
		return &abi.Event{}, fmt.Errorf("Log has no topics")
	}
	return app.contractABI.EventByID(log.Topics[0])
}

func (app *App) unpackEventData(event *abi.Event, log types.Log) (map[string]any, error) {
	args := map[string]any{}

	if err := app.contractABI.UnpackIntoMap(args, event.Name, log.Data); err != nil {
		return nil, fmt.Errorf("unpack data: %w", err)
	}
	return args, nil
}

func (app *App) unpackEventTopics(event *abi.Event, log types.Log) (map[string]any, error) {
	args := map[string]any{}

	var indexedInputs abi.Arguments

	for _, input := range event.Inputs {
		if input.Indexed {
			indexedInputs = append(indexedInputs, input)
		}
	}

	if err := abi.ParseTopicsIntoMap(args, indexedInputs, log.Topics[1:]); err != nil {
		return nil, fmt.Errorf("unpack topics: %w", err)
	}

	return args, nil
}

func mergeArgs(data map[string]any, topics map[string]any) map[string]any {
	maps.Copy(data, topics)
	return data
}

func buildVaultEvent(eventName string, merged map[string]any, log types.Log, timestamp time.Time) (VaultEvent, error) {
	baseEvent := VaultEvent{
		TxHash:      log.TxHash,
		LogIndex:    log.Index,
		BlockNumber: log.BlockNumber,
		TimeStamp:   timestamp,
		EventType:   eventName,
	}

	user, ok := merged["user"].(common.Address)
	if !ok {
		return VaultEvent{}, fmt.Errorf("missing/invalid field for event: %s", eventName)
	}
	baseEvent.WalletAddress = user

	switch eventName {
	case "UserDeposited", "UserWithdrawn", "UserRequestedWithdrawal":
		amount, ok := merged["amount"].(*big.Int)
		if !ok {
			return VaultEvent{}, fmt.Errorf("missing/invalid amount field for event: %s", eventName)
		}
		baseEvent.Amount = amount
	case "UserModifiedPendingWithdrawal":
		prevAmount, ok := merged["previousAmount"].(*big.Int)
		if !ok {
			return VaultEvent{}, fmt.Errorf("missing/invalid previous amount for event: %s", eventName)
		}
		newAmount, ok := merged["newAmount"].(*big.Int)
		if !ok {
			return VaultEvent{}, fmt.Errorf("missing/invalid previous amount for event: %s", eventName)
		}
		baseEvent.PreviousAmount = prevAmount
		baseEvent.NewAmount = newAmount
	default:
		return VaultEvent{}, fmt.Errorf("unhandled event type: %s", eventName)
	}
	return baseEvent, nil
}
