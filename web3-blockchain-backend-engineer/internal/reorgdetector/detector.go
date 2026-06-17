package reorgdetector

import (
	"context"
	"log/slog"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
	"github.com/nats-io/nats.go"
)

// ReorgDetector monitors blockchain for reorganizations and handles compensation
type ReorgDetector struct {
	chainID      int64
	rpcClient    rpcgateway.Client
	blockRepo    repository.BlockRepository
	depositRepo  repository.DepositRepository
	ledgerRepo   repository.LedgerEntryRepository
	natsClient   *nats.Conn
	logger       *slog.Logger

	// Detection settings
	checkInterval time.Duration // How often to run detection
	lookbackLimit int           // Number of blocks to look back for reorg detection
}

// ReorgEvent represents a detected chain reorganization event
type ReorgEvent struct {
	ChainID      int64  `json:"chain_id"`
	ForkBlock    uint64 `json:"fork_block"`
	NewBlockHash string `json:"new_block_hash"`
	OldBlockHash string `json:"old_block_hash"`
}

// NewReorgDetector creates a new ReorgDetector instance
func NewReorgDetector(
	chainID int64,
	rpcClient rpcgateway.Client,
	blockRepo repository.BlockRepository,
	depositRepo repository.DepositRepository,
	ledgerRepo repository.LedgerEntryRepository,
	natsClient *nats.Conn,
	logger *slog.Logger,
) *ReorgDetector {
	return &ReorgDetector{
		chainID:        chainID,
		rpcClient:      rpcClient,
		blockRepo:      blockRepo,
		depositRepo:    depositRepo,
		ledgerRepo:     ledgerRepo,
		natsClient:     natsClient,
		logger:         logger,
		checkInterval:  30 * time.Second,
		lookbackLimit:  10,
	}
}

// NewReorgDetectorWithOptions creates a ReorgDetector with custom options
func NewReorgDetectorWithOptions(
	chainID int64,
	rpcClient rpcgateway.Client,
	blockRepo repository.BlockRepository,
	depositRepo repository.DepositRepository,
	ledgerRepo repository.LedgerEntryRepository,
	natsClient *nats.Conn,
	logger *slog.Logger,
	checkInterval time.Duration,
	lookbackLimit int,
) *ReorgDetector {
	return &ReorgDetector{
		chainID:        chainID,
		rpcClient:      rpcClient,
		blockRepo:      blockRepo,
		depositRepo:    depositRepo,
		ledgerRepo:     ledgerRepo,
		natsClient:     natsClient,
		logger:         logger,
		checkInterval:  checkInterval,
		lookbackLimit:  lookbackLimit,
	}
}

// Start begins the continuous reorg detection loop
func (d *ReorgDetector) Start(ctx context.Context) {
	d.logger.Info("starting reorg detector",
		"chain_id", d.chainID,
		"check_interval", d.checkInterval,
		"lookback_limit", d.lookbackLimit,
	)

	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	// Run initial detection
	if err := d.Detect(ctx); err != nil {
		d.logger.Error("initial reorg detection failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("reorg detector stopped", "chain_id", d.chainID)
			return
		case <-ticker.C:
			if err := d.Detect(ctx); err != nil {
				d.logger.Error("reorg detection failed", "error", err)
			}
		}
	}
}

// Detect checks for chain reorganizations by comparing local blocks with on-chain data
func (d *ReorgDetector) Detect(ctx context.Context) error {
	// 1. Get local recent blocks (ordered from newest to oldest)
	localBlocks, err := d.blockRepo.ListByChainID(d.chainID, d.lookbackLimit)
	if err != nil {
		return err
	}

	if len(localBlocks) == 0 {
		d.logger.Debug("no local blocks found for detection", "chain_id", d.chainID)
		return nil
	}

	// Reverse to process from oldest to newest for proper chain continuity check
	for i, j := 0, len(localBlocks)-1; i < j; i, j = i+1, j-1 {
		localBlocks[i], localBlocks[j] = localBlocks[j], localBlocks[i]
	}

	// 2. Get the latest block number from chain
	latestBlockNum, err := d.rpcClient.BlockNumber(ctx)
	if err != nil {
		return err
	}

	d.logger.Debug("running reorg detection",
		"chain_id", d.chainID,
		"local_blocks", len(localBlocks),
		"latest_onchain_block", latestBlockNum,
	)

	// 3. Verify each local block's hash against on-chain data
	for i := 0; i < len(localBlocks); i++ {
		localBlock := localBlocks[i]

		// Skip blocks beyond latest on-chain height
		if uint64(localBlock.BlockNumber) > latestBlockNum {
			continue
		}

		// Get the corresponding on-chain block
		onChainBlock, err := d.rpcClient.GetBlockByNumber(ctx, uint64(localBlock.BlockNumber))
		if err != nil {
			d.logger.Warn("failed to fetch on-chain block, skipping",
				"block_number", localBlock.BlockNumber,
				"error", err,
			)
			continue
		}

		onChainHash := onChainBlock.Hash.Hex()

		// 4. Compare block hash - if mismatch, reorg detected
		if localBlock.BlockHash != onChainHash {
			d.logger.Warn("block hash mismatch - potential reorg detected",
				"chain_id", d.chainID,
				"block_number", localBlock.BlockNumber,
				"local_hash", localBlock.BlockHash,
				"onchain_hash", onChainHash,
			)

			// Handle the reorg starting from this block
			orphanedBlocks := localBlocks[i:]
			return d.handleReorg(ctx, orphanedBlocks, onChainHash)
		}

		// 5. Verify parent_hash continuity with next block
		if i < len(localBlocks)-1 {
			nextBlock := localBlocks[i+1]
			if nextBlock.ParentHash != localBlock.BlockHash {
				d.logger.Warn("parent hash discontinuity - reorg detected",
					"chain_id", d.chainID,
					"block_number", localBlock.BlockNumber,
					"local_parent_hash", localBlock.ParentHash,
					"expected_hash", nextBlock.BlockHash,
				)

				// Handle reorg starting from the next block
				orphanedBlocks := localBlocks[i+1:]
				return d.handleReorg(ctx, orphanedBlocks, onChainHash)
			}
		}
	}

	return nil
}

// DetectByBlockNumber checks for reorgs up to a specific block number
// This can be called after scanner saves new blocks
func (d *ReorgDetector) DetectByBlockNumber(ctx context.Context, fromBlockNum uint64) error {
	// Get local blocks from the specified block number
	localBlocks, err := d.blockRepo.ListByChainID(d.chainID, d.lookbackLimit)
	if err != nil {
		return err
	}

	if len(localBlocks) == 0 {
		return nil
	}

	// Filter blocks to only include those >= fromBlockNum
	var relevantBlocks []*repository.Block
	for _, block := range localBlocks {
		if uint64(block.BlockNumber) >= fromBlockNum {
			relevantBlocks = append(relevantBlocks, block)
		}
	}

	if len(relevantBlocks) == 0 {
		return nil
	}

	// Reverse to process from oldest to newest
	for i, j := 0, len(relevantBlocks)-1; i < j; i, j = i+1, j-1 {
		relevantBlocks[i], relevantBlocks[j] = relevantBlocks[j], relevantBlocks[i]
	}

	// Check each block
	for i := 0; i < len(relevantBlocks); i++ {
		localBlock := relevantBlocks[i]

		onChainBlock, err := d.rpcClient.GetBlockByNumber(ctx, uint64(localBlock.BlockNumber))
		if err != nil {
			continue
		}

		onChainHash := onChainBlock.Hash.Hex()

		if localBlock.BlockHash != onChainHash {
			orphanedBlocks := relevantBlocks[i:]
			return d.handleReorg(ctx, orphanedBlocks, onChainHash)
		}
	}

	return nil
}
