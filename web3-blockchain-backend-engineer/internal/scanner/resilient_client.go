package scanner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ResilientHealth exposes provider health to scanner callers.
type ResilientHealth = rpcgateway.HealthSnapshot

// HealthReporter is implemented by RPC clients that can expose provider health.
type HealthReporter interface {
	InspectProviders() []ResilientHealth
}

type resilientClient struct {
	rpc *rpcgateway.ResilientClient
}

// NewResilientClientFromRepo creates a multi-provider scanner RPC client backed by rpc_providers.
func NewResilientClientFromRepo(chainID int64, repo repository.RPCProviderRepository) (Client, error) {
	rc, err := rpcgateway.NewResilientClient(chainID, repo, rpcgateway.DefaultCallOptions())
	if err != nil {
		return nil, err
	}
	return &resilientClient{rpc: rc}, nil
}

func (c *resilientClient) BlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := c.rpc.Call(ctx, func(callCtx context.Context, client *ethclient.Client) error {
		n, err := client.BlockNumber(callCtx)
		if err != nil {
			return err
		}
		blockNumber = n
		return nil
	})
	return blockNumber, err
}

func (c *resilientClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*Block, error) {
	var result *Block
	err := c.rpc.Call(ctx, func(callCtx context.Context, client *ethclient.Client) error {
		block, err := client.BlockByNumber(callCtx, big.NewInt(int64(blockNumber)))
		if err != nil {
			return err
		}
		header := block.Header()
		result = &Block{
			Number:     header.Number.Uint64(),
			Hash:       header.Hash(),
			ParentHash: header.ParentHash,
			Time:       header.Time,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *resilientClient) GetLogsBatched(ctx context.Context, filter LogsFilter, batchSize uint64) ([]Log, error) {
	var allLogs []Log
	fromBlock := filter.FromBlock
	if fromBlock == nil {
		fromBlock = big.NewInt(0)
	}
	toBlock := filter.ToBlock
	if toBlock == nil {
		latest, err := c.BlockNumber(ctx)
		if err != nil {
			return nil, err
		}
		toBlock = new(big.Int).SetUint64(latest)
	}

	currentFrom := new(big.Int).Set(fromBlock)
	batchSizeBig := new(big.Int).SetUint64(batchSize)

	for currentFrom.Cmp(toBlock) <= 0 {
		currentTo := new(big.Int).Sub(new(big.Int).Add(currentFrom, batchSizeBig), big.NewInt(1))
		if currentTo.Cmp(toBlock) > 0 {
			currentTo = toBlock
		}

		logs, err := c.getLogs(ctx, LogsFilter{
			Address:   filter.Address,
			Topics:    filter.Topics,
			FromBlock: currentFrom,
			ToBlock:   currentTo,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get logs from block %d to %d: %w", currentFrom, currentTo, err)
		}
		allLogs = append(allLogs, logs...)
		currentFrom = new(big.Int).Add(currentTo, big.NewInt(1))
	}

	return allLogs, nil
}

func (c *resilientClient) getLogs(ctx context.Context, filter LogsFilter) ([]Log, error) {
	var result []Log
	err := c.rpc.Call(ctx, func(callCtx context.Context, client *ethclient.Client) error {
		query := ethereum.FilterQuery{
			FromBlock: filter.FromBlock,
			ToBlock:   filter.ToBlock,
			Topics:    [][]common.Hash{filter.Topics},
		}
		if filter.Address != (common.Address{}) {
			query.Addresses = []common.Address{filter.Address}
		}

		logs, err := client.FilterLogs(callCtx, query)
		if err != nil {
			return err
		}

		converted := make([]Log, len(logs))
		for i, l := range logs {
			converted[i] = Log{
				Address:     l.Address,
				Topics:      l.Topics,
				Data:        l.Data,
				BlockNumber: l.BlockNumber,
				TxHash:      l.TxHash,
				LogIndex:    l.Index,
				BlockHash:   l.BlockHash,
				Removed:     l.Removed,
			}
		}
		result = converted
		return nil
	})
	return result, err
}

func (c *resilientClient) InspectProviders() []ResilientHealth {
	return c.rpc.InspectProviders()
}
