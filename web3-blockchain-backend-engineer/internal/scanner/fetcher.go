package scanner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// Transfer event signature hash
var ERC20TransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// LogsFilter defines the filter criteria for fetching logs
type LogsFilter struct {
	Address   common.Address // Contract address to filter
	Topics    []common.Hash  // List of topic filters
	FromBlock *big.Int       // Start block (nil = earliest)
	ToBlock   *big.Int       // End block (nil = latest)
}

// Block represents basic block information
type Block struct {
	Number            uint64        `json:"number"`
	Hash              common.Hash   `json:"hash"`
	ParentHash        common.Hash   `json:"parentHash"`
	Time              uint64        `json:"timestamp"`
	TransactionHashes []common.Hash `json:"transactions"`
}

// Transaction represents a native ETH transfer candidate within a block.
type Transaction struct {
	Hash  common.Hash     `json:"hash"`
	From  common.Address  `json:"from"`
	To    *common.Address `json:"to"`
	Value *big.Int        `json:"value"`
}

// Log represents a contract log event
type Log struct {
	Address     common.Address `json:"address"`
	Topics      []common.Hash  `json:"topics"`
	Data        []byte         `json:"data"`
	BlockNumber uint64         `json:"blockNumber"`
	TxHash      common.Hash    `json:"transactionHash"`
	LogIndex    uint           `json:"logIndex"`
	BlockHash   common.Hash    `json:"blockHash"`
	Removed     bool           `json:"removed"`
}

// Client defines the interface for RPC client operations
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	GetLogsBatched(ctx context.Context, filter LogsFilter, batchSize uint64) ([]Log, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*Block, error)
	GetTransactionsByHashes(ctx context.Context, hashes []common.Hash) ([]Transaction, error)
}

// rpcClient implements Client using go-ethereum's rpc client
type rpcClient struct {
	client *rpc.Client
}

// NewRPCClient creates a new RPC client
func NewRPCClient(endpoint string) (Client, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC client: %w", err)
	}
	return &rpcClient{client: client}, nil
}

// BlockNumber returns the latest block number
func (c *rpcClient) BlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber hexUint64
	err := c.client.CallContext(ctx, &blockNumber, "eth_blockNumber")
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}
	return uint64(blockNumber), nil
}

// GetLogsBatched fetches logs in batches to avoid request size limits
func (c *rpcClient) GetLogsBatched(ctx context.Context, filter LogsFilter, batchSize uint64) ([]Log, error) {
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

		filterCopy := LogsFilter{
			Address:   filter.Address,
			Topics:    filter.Topics,
			FromBlock: currentFrom,
			ToBlock:   currentTo,
		}

		logs, err := c.getLogs(ctx, filterCopy)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs from block %d to %d: %w", currentFrom, currentTo, err)
		}

		allLogs = append(allLogs, logs...)

		currentFrom = new(big.Int).Add(currentTo, big.NewInt(1))
	}

	return allLogs, nil
}

// getLogs fetches logs for a single range
func (c *rpcClient) getLogs(ctx context.Context, filter LogsFilter) ([]Log, error) {
	filterArgs := map[string]interface{}{
		"fromBlock": toBlockParam(filter.FromBlock),
		"toBlock":   toBlockParam(filter.ToBlock),
		"topics":    filter.Topics,
	}

	if filter.Address != (common.Address{}) {
		filterArgs["address"] = filter.Address
	}

	var rawLogs []rpcLog
	err := c.client.CallContext(ctx, &rawLogs, "eth_getLogs", filterArgs)
	if err != nil {
		return nil, err
	}

	logs := make([]Log, len(rawLogs))
	for i, raw := range rawLogs {
		logs[i] = Log{
			Address:     raw.Address,
			Topics:      raw.Topics,
			Data:        common.FromHex(raw.Data),
			BlockNumber: uint64(raw.BlockNumber),
			TxHash:      raw.TxHash,
			LogIndex:    uint(raw.LogIndex),
			BlockHash:   raw.BlockHash,
			Removed:     raw.Removed,
		}
	}

	return logs, nil
}

func toBlockParam(block *big.Int) interface{} {
	if block == nil {
		return nil
	}
	if block.Sign() < 0 {
		return nil
	}
	return fmt.Sprintf("0x%x", block.Uint64())
}

// GetBlockByNumber returns block information for the given block number
func (c *rpcClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*Block, error) {
	var raw blockRaw
	blockNumHex := toHexUint64(blockNumber)
	err := c.client.CallContext(ctx, &raw, "eth_getBlockByNumber", blockNumHex, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	return &Block{
		Number:            uint64(raw.Number),
		Hash:              raw.Hash,
		ParentHash:        raw.ParentHash,
		Time:              uint64(raw.Timestamp),
		TransactionHashes: raw.TransactionHashes,
	}, nil
}

func (c *rpcClient) GetTransactionsByHashes(ctx context.Context, hashes []common.Hash) ([]Transaction, error) {
	if len(hashes) == 0 {
		return nil, nil
	}

	transactions := make([]Transaction, 0, len(hashes))
	for start := 0; start < len(hashes); start += 100 {
		end := start + 100
		if end > len(hashes) {
			end = len(hashes)
		}

		chunk, err := c.getTransactionsByHashesChunk(ctx, hashes[start:end])
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, chunk...)
	}

	return transactions, nil
}

func (c *rpcClient) getTransactionsByHashesChunk(ctx context.Context, hashes []common.Hash) ([]Transaction, error) {
	results := make([]*transactionRaw, len(hashes))
	batch := make([]rpc.BatchElem, len(hashes))
	for i, hash := range hashes {
		batch[i] = rpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{hash},
			Result: &results[i],
		}
	}

	if err := c.client.BatchCallContext(ctx, batch); err != nil {
		return nil, fmt.Errorf("failed to batch fetch transactions: %w", err)
	}

	transactions := make([]Transaction, 0, len(results))
	for i := range batch {
		if batch[i].Error != nil {
			return nil, fmt.Errorf("failed to fetch transaction %s: %w", hashes[i].Hex(), batch[i].Error)
		}
		if results[i] == nil {
			continue
		}
		transactions = append(transactions, results[i].toTransaction())
	}

	return transactions, nil
}

// eth_getBlockByNumber returns a block with number as hex string
type blockRaw struct {
	Number            hexUint64     `json:"number"`
	Hash              common.Hash   `json:"hash"`
	ParentHash        common.Hash   `json:"parentHash"`
	Timestamp         hexUint64     `json:"timestamp"`
	TransactionHashes []common.Hash `json:"transactions"`
}

type transactionRaw struct {
	Hash  common.Hash     `json:"hash"`
	From  common.Address  `json:"from"`
	To    *common.Address `json:"to"`
	Value *hexutil.Big    `json:"value"`
}

func (t *transactionRaw) toTransaction() Transaction {
	value := big.NewInt(0)
	if t.Value != nil {
		value = (*big.Int)(t.Value)
	}
	return Transaction{
		Hash:  t.Hash,
		From:  t.From,
		To:    t.To,
		Value: new(big.Int).Set(value),
	}
}

type hexUint64 uint64

func (h *hexUint64) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == "\"0x0\"" {
		*h = 0
		return nil
	}
	// Remove quotes if present
	str = trimQuotes(str)
	val, err := parseHexUint64(str)
	if err != nil {
		return err
	}
	*h = hexUint64(val)
	return nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseHexUint64(s string) (uint64, error) {
	if len(s) > 2 && s[:2] == "0x" {
		var val uint64
		for _, c := range s[2:] {
			val <<= 4
			switch {
			case c >= '0' && c <= '9':
				val |= uint64(c - '0')
			case c >= 'a' && c <= 'f':
				val |= uint64(c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				val |= uint64(c - 'A' + 10)
			default:
				return 0, fmt.Errorf("invalid hex character: %c", c)
			}
		}
		return val, nil
	}
	return 0, fmt.Errorf("invalid hex string: %s", s)
}

func toHexUint64(n uint64) string {
	return fmt.Sprintf("0x%x", n)
}

// rpcLog represents the raw log from JSON-RPC
type rpcLog struct {
	Address     common.Address `json:"address"`
	Topics      []common.Hash  `json:"topics"`
	Data        string         `json:"data"`
	BlockNumber hexUint64      `json:"blockNumber"`
	TxHash      common.Hash    `json:"transactionHash"`
	LogIndex    hexUint64      `json:"logIndex"`
	BlockHash   common.Hash    `json:"blockHash"`
	Removed     bool           `json:"removed"`
}

//go:generate mockgen -destination=mock_rpcclient_test.go -package=scanner . Client
//go:generate mockgen -destination=mock_checkpoint_repo_test.go -package=scanner github.com/asset-platform/multi-chain-asset-platform/internal/repository ScanCheckpointRepository
//go:generate mockgen -destination=mock_block_repo_test.go -package=scanner github.com/asset-platform/multi-chain-asset-platform/internal/repository BlockRepository
