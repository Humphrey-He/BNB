package rpcgateway

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Block represents a simplified block structure for RPC responses
type Block struct {
	Number     uint64      `json:"number"`
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
	Timestamp  uint64      `json:"timestamp"`
}

// Transaction represents a simplified transaction structure
type Transaction struct {
	Hash      common.Address `json:"hash"`
	To        *common.Address `json:"to"`
	Value     *big.Int       `json:"value"`
	Gas       uint64        `json:"gas"`
	GasPrice  *big.Int      `json:"gasPrice"`
	Input     []byte        `json:"input"`
	Nonce     uint64        `json:"nonce"`
	ChainID   *big.Int      `json:"chainId"`
}

// Log represents a simplified log entry
type Log struct {
	Address     common.Address `json:"address"`
	Topics      []common.Hash `json:"topics"`
	Data        []byte        `json:"data"`
	BlockNumber uint64        `json:"blockNumber"`
	TxHash      common.Hash   `json:"transactionHash"`
	TxIndex     uint          `json:"transactionIndex"`
	BlockHash   common.Hash   `json:"blockHash"`
	LogIndex    uint          `json:"logIndex"`
	Removed     bool          `json:"removed"`
}

// Receipt represents a simplified transaction receipt
type Receipt struct {
	TxHash      common.Hash `json:"transactionHash"`
	BlockNumber *big.Int    `json:"blockNumber"`
	BlockHash   common.Hash `json:"blockHash"`
	GasUsed     uint64      `json:"gasUsed"`
	Status      uint64      `json:"status"`
	Logs        []Log      `json:"logs"`
}

// Client interface defines the RPC client operations
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, num uint64) (*Block, error)
	GetLogs(ctx context.Context, filter LogsFilter) ([]Log, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*Receipt, error)
}

// rpcClient implements Client interface using go-ethereum
type rpcClient struct {
	provider  *Provider
	ethClient *ethclient.Client
}

// NewClient creates a new RPC client for a provider
func NewClient(provider *Provider) (Client, error) {
	client, err := ethclient.Dial(provider.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC provider %s: %w", provider.Name, err)
	}

	return &rpcClient{
		provider:  provider,
		ethClient: client,
	}, nil
}

// BlockNumber returns the current block number
func (c *rpcClient) BlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("BlockNumber call failed: %w", err)
	}
	return blockNumber, nil
}

// GetBlockByNumber returns the block with the given number
func (c *rpcClient) GetBlockByNumber(ctx context.Context, num uint64) (*Block, error) {
	block, err := c.ethClient.BlockByNumber(ctx, big.NewInt(int64(num)))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber call failed for block %d: %w", num, err)
	}

	header := block.Header()
	return &Block{
		Number:     header.Number.Uint64(),
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Timestamp:  header.Time,
	}, nil
}

// GetLogs returns logs matching the given filter
func (c *rpcClient) GetLogs(ctx context.Context, filter LogsFilter) ([]Log, error) {
	ethFilter := ethereum.FilterQuery{
		Addresses: []common.Address{filter.Address},
		Topics:    [][]common.Hash{filter.Topics},
		FromBlock: big.NewInt(int64(filter.FromBlock)),
		ToBlock:   big.NewInt(int64(filter.ToBlock)),
	}

	logs, err := c.ethClient.FilterLogs(ctx, ethFilter)
	if err != nil {
		return nil, fmt.Errorf("GetLogs call failed: %w", err)
	}

	result := make([]Log, len(logs))
	for i, l := range logs {
		result[i] = Log{
			Address:     l.Address,
			Topics:      l.Topics,
			Data:        l.Data,
			BlockNumber: l.BlockNumber,
			TxHash:      l.TxHash,
			TxIndex:     l.TxIndex,
			BlockHash:   l.BlockHash,
			LogIndex:    l.Index,
			Removed:     l.Removed,
		}
	}
	return result, nil
}

// GetTransactionReceipt returns the receipt for the given transaction hash
func (c *rpcClient) GetTransactionReceipt(ctx context.Context, txHash string) (*Receipt, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.ethClient.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("GetTransactionReceipt call failed for tx %s: %w", txHash, err)
	}

	logs := make([]Log, len(receipt.Logs))
	for i, l := range receipt.Logs {
		logs[i] = Log{
			Address:     l.Address,
			Topics:      l.Topics,
			Data:        l.Data,
			BlockNumber: l.BlockNumber,
			TxHash:      l.TxHash,
			TxIndex:     l.TxIndex,
			BlockHash:   l.BlockHash,
			LogIndex:    l.Index,
			Removed:     l.Removed,
		}
	}

	return &Receipt{
		TxHash:      receipt.TxHash,
		BlockNumber: receipt.BlockNumber,
		BlockHash:   receipt.BlockHash,
		GasUsed:     receipt.GasUsed,
		Status:      receipt.Status,
		Logs:        logs,
	}, nil
}
