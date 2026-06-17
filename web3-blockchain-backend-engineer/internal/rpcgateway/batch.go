package rpcgateway

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// LogsFilter represents the filter criteria for eth_getLogs
type LogsFilter struct {
	Address   common.Address `json:"address"`
	Topics    []common.Hash  `json:"topics"`
	FromBlock uint64         `json:"fromBlock"`
	ToBlock   uint64         `json:"toBlock"`
}

// DefaultBatchSize is the default batch size for log queries
const DefaultBatchSize uint64 = 1000

// MaxBatchSize is the maximum allowed batch size
const MaxBatchSize uint64 = 10000

// ERC-20 Transfer event signature
var ERC20TransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// ERC-20 Approval event signature
var ERC20ApprovalTopic = common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")

// ERC-20 Transfer event topics
var (
	ERC20TransferEventTopic0 = ERC20TransferTopic
	ERC20TransferEventTopic1 = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000") // Anonymous
)

// BatchConfig holds configuration for batch operations
type BatchConfig struct {
	BatchSize      uint64        // Number of blocks per batch
	Concurrency    int           // Number of concurrent requests
	RetryAttempts  int           // Number of retry attempts per batch
	RetryDelay     time.Duration // Delay between retries
	RequestTimeout time.Duration // Timeout for each request
}

// DefaultBatchConfig returns the default batch configuration
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		BatchSize:      DefaultBatchSize,
		Concurrency:    4,
		RetryAttempts:  3,
		RetryDelay:     500 * time.Millisecond,
		RequestTimeout: 30 * time.Second,
	}
}

// GetLogsBatcher handles batched log retrieval to avoid timeout issues
type GetLogsBatcher struct {
	client  Client
	config  BatchConfig
	mu      sync.Mutex
}

// NewGetLogsBatcher creates a new GetLogsBatcher
func NewGetLogsBatcher(client Client, config BatchConfig) *GetLogsBatcher {
	if config.BatchSize == 0 {
		config.BatchSize = DefaultBatchSize
	}
	if config.BatchSize > MaxBatchSize {
		config.BatchSize = MaxBatchSize
	}
	if config.Concurrency == 0 {
		config.Concurrency = 4
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	return &GetLogsBatcher{
		client: client,
		config: config,
	}
}

// GetLogsBatched retrieves logs in batches to avoid RPC timeout
func (b *GetLogsBatcher) GetLogsBatched(ctx context.Context, filter LogsFilter) ([]Log, error) {
	if filter.FromBlock > filter.ToBlock {
		return nil, ErrInvalidBlockRange
	}

	// Calculate total blocks
	totalBlocks := filter.ToBlock - filter.FromBlock + 1

	// If total blocks is small enough, do a single request
	if totalBlocks <= b.config.BatchSize {
		return b.client.GetLogs(ctx, filter)
	}

	// Otherwise, batch the requests
	return b.getLogsInBatches(ctx, filter)
}

// getLogsInBatches retrieves logs in multiple batched requests
func (b *GetLogsBatcher) getLogsInBatches(ctx context.Context, filter LogsFilter) ([]Log, error) {
	var allLogs []Log
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error
	errChan := make(chan error, b.config.Concurrency)

	// Create a channel for batches
	batchChan := make(chan LogsFilter, b.config.Concurrency)

	// Start workers
	for i := 0; i < b.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				logs, err := b.fetchLogsWithRetry(ctx, batch)
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}

				mu.Lock()
				allLogs = append(allLogs, logs...)
				mu.Unlock()
			}
		}()
	}

	// Send batches
	currentBlock := filter.FromBlock
	for currentBlock <= filter.ToBlock {
		endBlock := currentBlock + b.config.BatchSize - 1
		if endBlock > filter.ToBlock {
			endBlock = filter.ToBlock
		}

		batchFilter := LogsFilter{
			Address:   filter.Address,
			Topics:    filter.Topics,
			FromBlock: currentBlock,
			ToBlock:   endBlock,
		}

		select {
		case batchChan <- batchFilter:
		case <-ctx.Done():
			close(batchChan)
			return nil, ctx.Err()
		}

		currentBlock = endBlock + 1
	}

	close(batchChan)
	wg.Wait()

	// Check for errors
	select {
	case err := <-errChan:
		return nil, err
	default:
		if firstErr != nil {
			return nil, firstErr
		}
	}

	// Sort logs by block number and log index
	sortLogs(allLogs)

	return allLogs, nil
}

// fetchLogsWithRetry fetches logs with retry logic
func (b *GetLogsBatcher) fetchLogsWithRetry(ctx context.Context, filter LogsFilter) ([]Log, error) {
	var lastErr error

	for attempt := 0; attempt <= b.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(b.config.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Create a timeout context for this request
		reqCtx, cancel := context.WithTimeout(ctx, b.config.RequestTimeout)
		defer cancel()

		logs, err := b.client.GetLogs(reqCtx, filter)
		if err == nil {
			return logs, nil
		}

		lastErr = err

		// Check if context was cancelled
		if reqCtx.Err() == context.Canceled {
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed to fetch logs after %d attempts: %w", b.config.RetryAttempts, lastErr)
}

// CalculateBlockRanges calculates the block ranges for batch processing
func CalculateBlockRanges(fromBlock, toBlock, batchSize uint64) []LogsFilter {
	var ranges []LogsFilter

	currentBlock := fromBlock
	for currentBlock <= toBlock {
		endBlock := currentBlock + batchSize - 1
		if endBlock > toBlock {
			endBlock = toBlock
		}

		ranges = append(ranges, LogsFilter{
			FromBlock: currentBlock,
			ToBlock:   endBlock,
		})

		currentBlock = endBlock + 1
	}

	return ranges
}

// GetLogsByRange retrieves logs within a specific block range
// This is useful for parallel processing of different ranges
func (b *GetLogsBatcher) GetLogsByRange(ctx context.Context, filter LogsFilter, batchSize uint64) ([]Log, error) {
	if batchSize > MaxBatchSize {
		return nil, ErrBatchSizeTooLarge
	}

	batcher := NewGetLogsBatcher(b.client, BatchConfig{
		BatchSize:      batchSize,
		Concurrency:    b.config.Concurrency,
		RetryAttempts:  b.config.RetryAttempts,
		RetryDelay:     b.config.RetryDelay,
		RequestTimeout: b.config.RequestTimeout,
	})

	return batcher.GetLogsBatched(ctx, filter)
}

// LogBatch represents a batch of logs for processing
type LogBatch struct {
	Logs      []Log
	FromBlock uint64
	ToBlock   uint64
}

// ProcessLogsInBatches processes logs in batches with a handler function
func (b *GetLogsBatcher) ProcessLogsInBatches(ctx context.Context, filter LogsFilter, handler func([]Log) error) error {
	if filter.FromBlock > filter.ToBlock {
		return ErrInvalidBlockRange
	}

	totalBlocks := filter.ToBlock - filter.FromBlock + 1
	batches := (totalBlocks + b.config.BatchSize - 1) / b.config.BatchSize

	var wg sync.WaitGroup
	errChan := make(chan error, batches)
	resultChan := make(chan []Log, b.config.Concurrency)

	// Process batches concurrently
	for start := filter.FromBlock; start <= filter.ToBlock; {
		end := start + b.config.BatchSize - 1
		if end > filter.ToBlock {
			end = filter.ToBlock
		}

		batchFilter := LogsFilter{
			Address:   filter.Address,
			Topics:    filter.Topics,
			FromBlock: start,
			ToBlock:   end,
		}

		wg.Add(1)
		go func(f LogsFilter) {
			defer wg.Done()

			logs, err := b.GetLogsBatched(ctx, f)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			if len(logs) > 0 {
				select {
				case resultChan <- logs:
				default:
				}
			}
		}(batchFilter)

		start = end + 1
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Process results as they come in
	for logs := range resultChan {
		if err := handler(logs); err != nil {
			return err
		}
	}

	// Check for errors
	if err, ok := <-errChan; ok {
		return err
	}

	return nil
}

// sortLogs sorts logs by block number and log index
func sortLogs(logs []Log) {
	if len(logs) <= 1 {
		return
	}

	// Use a simple bubble sort since the list is likely small
	n := len(logs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if logs[j].BlockNumber > logs[j+1].BlockNumber ||
				(logs[j].BlockNumber == logs[j+1].BlockNumber && logs[j].LogIndex > logs[j+1].LogIndex) {
				logs[j], logs[j+1] = logs[j+1], logs[j]
			}
		}
	}
}

// EstimateBatches estimates the number of batches needed for a block range
func EstimateBatches(fromBlock, toBlock, batchSize uint64) int {
	if fromBlock > toBlock {
		return 0
	}
	blocks := toBlock - fromBlock + 1
	return int((blocks + batchSize - 1) / batchSize)
}

// GetOptimalBatchSize calculates the optimal batch size based on expected results
func GetOptimalBatchSize(fromBlock, toBlock uint64, expectedLogsPerBlock float64) uint64 {
	// Aim for batches that return around 1000-5000 logs
	targetLogsPerBatch := 2000.0

	if expectedLogsPerBlock <= 0 {
		return DefaultBatchSize
	}

	optimalBatchSize := uint64(math.Ceil(targetLogsPerBatch / expectedLogsPerBlock))

	if optimalBatchSize < 100 {
		optimalBatchSize = 100
	}
	if optimalBatchSize > MaxBatchSize {
		optimalBatchSize = MaxBatchSize
	}

	return optimalBatchSize
}
