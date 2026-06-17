package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// Config holds scanner configuration
type Config struct {
	ChainID      int64
	BatchSize    uint64
	PollInterval time.Duration
	NATSSubject  string
}

// DefaultConfig returns the default scanner configuration
func DefaultConfig(chainID int64) Config {
	return Config{
		ChainID:      chainID,
		BatchSize:    100,
		PollInterval: time.Second,
		NATSSubject:  "raw_events",
	}
}

// RawEventMessage represents the message published to NATS
type RawEventMessage struct {
	ChainID        int64    `json:"chain_id"`
	BlockNumber    uint64   `json:"block_number"`
	BlockHash      string   `json:"block_hash"`
	TxHash         string   `json:"tx_hash"`
	LogIndex       uint     `json:"log_index"`
	ContractAddress string   `json:"contract_address"`
	Topics         []string `json:"topics"`
	Data           string   `json:"data"`
}

// Scanner is the main scanner service that scans blockchain for events
type Scanner struct {
	chainID    int64
	rpcClient  Client
	checkpoint repository.ScanCheckpointRepository
	blockRepo  repository.BlockRepository
	natsClient *nats.Conn
	logger     *slog.Logger
	persistence *persistence
	metrics    *Metrics
	config     Config

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewScanner creates a new Scanner instance
func NewScanner(
	chainID int64,
	rpcClient Client,
	checkpointRepo repository.ScanCheckpointRepository,
	blockRepo repository.BlockRepository,
	natsClient *nats.Conn,
	logger *slog.Logger,
	metrics *Metrics,
	config Config,
) *Scanner {
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.PollInterval == 0 {
		config.PollInterval = time.Second
	}
	if config.NATSSubject == "" {
		config.NATSSubject = "raw_events"
	}

	persistence := newPersistence(checkpointRepo, blockRepo, chainID, logger)

	return &Scanner{
		chainID:    chainID,
		rpcClient:  rpcClient,
		checkpoint: checkpointRepo,
		blockRepo:  blockRepo,
		natsClient: natsClient,
		logger:     logger,
		persistence: persistence,
		metrics:    metrics,
		config:     config,
	}
}

// Run starts the scanner main loop
func (s *Scanner) Run(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.wg.Wait()

	// Ensure checkpoint exists
	_, err := s.persistence.EnsureCheckpoint(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure checkpoint: %w", err)
	}

	s.metrics.SetBatchSize(s.config.BatchSize)
	s.logger.Info("scanner started",
		"chain_id", s.chainID,
		"batch_size", s.config.BatchSize,
		"poll_interval", s.config.PollInterval,
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scanner shutting down")
			return nil
		default:
			if err := s.scanLoop(ctx); err != nil {
				s.metrics.IncErrors()
				s.logger.Error("scan loop error", "error", err)
				// Wait before retrying on error
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
					continue
				}
			}
		}
	}
}

// scanLoop performs one iteration of the scan loop
func (s *Scanner) scanLoop(ctx context.Context) error {
	// Create a timeout context for this batch
	batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 1. Get current chain head block
	latestBlock, err := s.rpcClient.BlockNumber(batchCtx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// 2. Get local checkpoint
	checkpoint, err := s.persistence.GetCheckpoint(batchCtx)
	if err != nil {
		return fmt.Errorf("failed to get checkpoint: %w", err)
	}

	// 3. Calculate scan range
	fromBlock := checkpoint.LastScannedBlock + 1
	toBlock := int64(min(latestBlock, uint64(fromBlock)+s.config.BatchSize-1))

	// Update scan lag metric
	s.metrics.SetScanLag(latestBlock - uint64(fromBlock))

	// Check if there's anything to scan
	if toBlock < fromBlock {
		// No new blocks, wait and retry
		s.logger.Debug("no new blocks to scan",
			"latest_block", latestBlock,
			"last_scanned", checkpoint.LastScannedBlock,
		)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.config.PollInterval):
			return nil
		}
	}

	s.logger.Info("scanning blocks",
		"from_block", fromBlock,
		"to_block", toBlock,
		"batch_size", toBlock - fromBlock + 1,
	)

	// 4. Build filter and fetch logs
	filter := LogsFilter{
		Topics:    TransferFilter.Topics,
		FromBlock: int64ToBigInt(fromBlock),
		ToBlock:   int64ToBigInt(toBlock),
	}

	logs, err := s.rpcClient.GetLogsBatched(batchCtx, filter, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	s.logger.Info("fetched logs", "count", len(logs))

	// 5. Save ALL blocks in the scanned range for reorg detection (not just blocks with events)
	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		block, err := s.rpcClient.GetBlockByNumber(batchCtx, uint64(blockNum))
		if err != nil {
			s.logger.Warn("failed to fetch block for persistence", "block_number", blockNum, "error", err)
			continue
		}
		if err := s.persistence.SaveBlock(batchCtx, block); err != nil {
			s.logger.Warn("failed to save block", "block_number", blockNum, "error", err)
		}
	}

	// 6. Publish logs to NATS
	if err := s.publishLogs(batchCtx, logs); err != nil {
		return fmt.Errorf("failed to publish logs: %w", err)
	}

	// 7. Update checkpoint
	if err := s.persistence.UpdateCheckpoint(batchCtx, toBlock); err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}

	// 8. Update metrics
	s.metrics.IncBlocksScanned(uint64(toBlock - fromBlock + 1))
	s.metrics.IncEventsExtracted(uint64(len(logs)))
	s.metrics.SetLastScannedBlock(uint64(toBlock))

	s.logger.Info("batch completed",
		"blocks_scanned", toBlock-fromBlock+1,
		"events_extracted", len(logs),
		"last_scanned_block", toBlock,
	)

	return nil
}

// publishLogs publishes log events to NATS
func (s *Scanner) publishLogs(ctx context.Context, logs []Log) error {
	if s.natsClient == nil {
		s.logger.Warn("NATS client not configured, skipping publish")
		return nil
	}

	for _, log := range logs {
		topics := make([]string, len(log.Topics))
		for i, topic := range log.Topics {
			topics[i] = topic.Hex()
		}

		msg := RawEventMessage{
			ChainID:        s.chainID,
			BlockNumber:    log.BlockNumber,
			BlockHash:      log.BlockHash.Hex(),
			TxHash:         log.TxHash.Hex(),
			LogIndex:       log.LogIndex,
			ContractAddress: log.Address.Hex(),
			Topics:         topics,
			Data:           bytesToHex(log.Data),
		}

		data, err := json.Marshal(msg)
		if err != nil {
			s.metrics.IncNatsPublishError()
			s.logger.Warn("failed to marshal log message", "error", err)
			continue
		}

		if err := s.natsClient.Publish(s.config.NATSSubject, data); err != nil {
			s.metrics.IncNatsPublishError()
			s.logger.Warn("failed to publish log", "error", err)
			continue
		}

		s.metrics.IncNatsPublish(1)
	}

	return nil
}

// Stop gracefully stops the scanner
func (s *Scanner) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.logger.Info("scanner stopped")
}

// Helper functions

func int64ToBigInt(n int64) *big.Int {
	return new(big.Int).SetInt64(n)
}

func bytesToHex(data []byte) string {
	if len(data) == 0 {
		return "0x"
	}
	return "0x" + hexEncode(data)
}

func hexEncode(data []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for i, b := range data {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
