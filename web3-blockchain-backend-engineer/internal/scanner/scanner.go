package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/ethereum/go-ethereum/common"
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
	ChainID         int64    `json:"chain_id"`
	BlockNumber     uint64   `json:"block_number"`
	BlockHash       string   `json:"block_hash"`
	TxHash          string   `json:"tx_hash"`
	LogIndex        uint     `json:"log_index"`
	ContractAddress string   `json:"contract_address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	EventName       string   `json:"event_name,omitempty"`
	FromAddress     string   `json:"from_address,omitempty"`
	ToAddress       string   `json:"to_address,omitempty"`
	Value           string   `json:"value,omitempty"`
}

// Scanner is the main scanner service that scans blockchain for events
type Scanner struct {
	chainID         int64
	rpcClient       Client
	checkpoint      repository.ScanCheckpointRepository
	blockRepo       repository.BlockRepository
	watchedAddrRepo repository.WatchedAddressRepository
	natsClient      *nats.Conn
	logger          *slog.Logger
	persistence     *persistence
	metrics         *Metrics
	config          Config

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewScanner creates a new Scanner instance
func NewScanner(
	chainID int64,
	rpcClient Client,
	checkpointRepo repository.ScanCheckpointRepository,
	blockRepo repository.BlockRepository,
	watchedAddrRepo repository.WatchedAddressRepository,
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
		chainID:         chainID,
		rpcClient:       rpcClient,
		checkpoint:      checkpointRepo,
		blockRepo:       blockRepo,
		watchedAddrRepo: watchedAddrRepo,
		natsClient:      natsClient,
		logger:          logger,
		persistence:     persistence,
		metrics:         metrics,
		config:          config,
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
	batchCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
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
		"batch_size", toBlock-fromBlock+1,
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

	rawEvents := s.buildRawLogEvents(logs)
	watchedAddresses, err := s.loadWatchedAddresses()
	if err != nil {
		return fmt.Errorf("failed to load watched addresses: %w", err)
	}

	// 5. Save ALL blocks in the scanned range for reorg detection (not just blocks with events)
	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		block, err := s.rpcClient.GetBlockByNumber(batchCtx, uint64(blockNum))
		if err != nil {
			return fmt.Errorf("failed to fetch block metadata for block %d: %w", blockNum, err)
		}
		if err := s.persistence.SaveBlock(batchCtx, block); err != nil {
			s.logger.Warn("failed to save block", "block_number", blockNum, "error", err)
		}
		nativeEvents, err := s.buildNativeTransferEvents(batchCtx, block, watchedAddresses)
		if err != nil {
			return fmt.Errorf("failed to build native transfer events for block %d: %w", blockNum, err)
		}
		rawEvents = append(rawEvents, nativeEvents...)
	}

	// 6. Publish logs to NATS
	if err := s.publishRawEvents(batchCtx, rawEvents); err != nil {
		return fmt.Errorf("failed to publish raw events: %w", err)
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
func (s *Scanner) publishRawEvents(ctx context.Context, events []RawEventMessage) error {
	if s.natsClient == nil {
		s.logger.Warn("NATS client not configured, skipping publish")
		return nil
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			s.metrics.IncNatsPublishError()
			s.logger.Warn("failed to marshal raw event", "error", err)
			continue
		}

		if err := s.natsClient.Publish(s.config.NATSSubject, data); err != nil {
			s.metrics.IncNatsPublishError()
			s.logger.Warn("failed to publish raw event", "error", err)
			continue
		}

		s.metrics.IncNatsPublish(1)
	}

	return nil
}

func (s *Scanner) buildRawLogEvents(logs []Log) []RawEventMessage {
	if len(logs) == 0 {
		return nil
	}

	events := make([]RawEventMessage, 0, len(logs))
	for _, log := range logs {
		topics := make([]string, len(log.Topics))
		for i, topic := range log.Topics {
			topics[i] = topic.Hex()
		}

		events = append(events, RawEventMessage{
			ChainID:         s.chainID,
			BlockNumber:     log.BlockNumber,
			BlockHash:       log.BlockHash.Hex(),
			TxHash:          log.TxHash.Hex(),
			LogIndex:        log.LogIndex,
			ContractAddress: log.Address.Hex(),
			Topics:          topics,
			Data:            bytesToHex(log.Data),
			EventName:       "Transfer",
		})
	}
	return events
}

func (s *Scanner) buildNativeTransferEvents(
	ctx context.Context,
	block *Block,
	watchedAddresses map[string]struct{},
) ([]RawEventMessage, error) {
	if block == nil || len(block.TransactionHashes) == 0 || len(watchedAddresses) == 0 {
		return nil, nil
	}

	transactions, err := s.rpcClient.GetTransactionsByHashes(ctx, block.TransactionHashes)
	if err != nil {
		return nil, err
	}

	events := make([]RawEventMessage, 0, len(transactions))
	for _, tx := range transactions {
		if tx.To == nil || tx.Value == nil || tx.Value.Sign() <= 0 {
			continue
		}
		if _, ok := watchedAddresses[strings.ToLower(tx.To.Hex())]; !ok {
			continue
		}

		events = append(events, RawEventMessage{
			ChainID:         s.chainID,
			BlockNumber:     block.Number,
			BlockHash:       block.Hash.Hex(),
			TxHash:          tx.Hash.Hex(),
			LogIndex:        0,
			ContractAddress: "0x0000000000000000000000000000000000000000",
			Topics:          []string{},
			Data:            "0x",
			EventName:       "NativeTransfer",
			FromAddress:     tx.From.Hex(),
			ToAddress:       tx.To.Hex(),
			Value:           tx.Value.String(),
		})
	}
	return events, nil
}

func (s *Scanner) loadWatchedAddresses() (map[string]struct{}, error) {
	if s.watchedAddrRepo == nil {
		return nil, nil
	}

	addresses, err := s.watchedAddrRepo.ListByChainID(s.chainID)
	if err != nil {
		return nil, err
	}

	watched := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		if address == nil || !address.IsActive || !common.IsHexAddress(address.Address) {
			continue
		}
		watched[strings.ToLower(address.Address)] = struct{}{}
	}
	return watched, nil
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
