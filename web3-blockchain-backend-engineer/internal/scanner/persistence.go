package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/ethereum/go-ethereum/common"
)

// TransferFilter defines the filter for ERC-20 Transfer events
var TransferFilter = LogsFilter{
	Topics: []common.Hash{
		ERC20TransferTopic, // Transfer(address,address,uint256)
	},
}

// persistence handles block persistence and checkpoint management
type persistence struct {
	checkpointRepo repository.ScanCheckpointRepository
	blockRepo      repository.BlockRepository
	chainID        int64
	logger         *slog.Logger
}

// newPersistence creates a new persistence handler
func newPersistence(
	checkpointRepo repository.ScanCheckpointRepository,
	blockRepo repository.BlockRepository,
	chainID int64,
	logger *slog.Logger,
) *persistence {
	return &persistence{
		checkpointRepo: checkpointRepo,
		blockRepo:      blockRepo,
		chainID:        chainID,
		logger:         logger,
	}
}

// GetCheckpoint retrieves the scan checkpoint for the chain
func (p *persistence) GetCheckpoint(ctx context.Context) (*repository.ScanCheckpoint, error) {
	checkpoint, err := p.checkpointRepo.GetByChainID(p.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint for chain %d: %w", p.chainID, err)
	}
	return checkpoint, nil
}

// EnsureCheckpoint ensures a checkpoint exists for the chain
func (p *persistence) EnsureCheckpoint(ctx context.Context) (*repository.ScanCheckpoint, error) {
	checkpoint, err := p.checkpointRepo.GetByChainID(p.chainID)
	if err != nil {
		// Create new checkpoint starting from block 0
		checkpoint = &repository.ScanCheckpoint{
			ChainID:          p.chainID,
			LastScannedBlock: 0,
			LastScannedAt:    time.Now(),
		}
		if createErr := p.checkpointRepo.Create(checkpoint); createErr != nil {
			return nil, fmt.Errorf("failed to create checkpoint for chain %d: %w", p.chainID, createErr)
		}
		p.logger.Info("created new checkpoint", "chain_id", p.chainID, "start_block", 0)
	}
	return checkpoint, nil
}

// UpdateCheckpoint updates the last scanned block for the chain
func (p *persistence) UpdateCheckpoint(ctx context.Context, blockNumber int64) error {
	err := p.checkpointRepo.UpdateLastScannedBlock(p.chainID, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to update checkpoint for chain %d at block %d: %w", p.chainID, blockNumber, err)
	}
	p.logger.Debug("checkpoint updated", "chain_id", p.chainID, "block_number", blockNumber)
	return nil
}

// SaveBlock saves block information for reorg detection
func (p *persistence) SaveBlock(ctx context.Context, block *Block) error {
	dbBlock := &repository.Block{
		ChainID:       p.chainID,
		BlockNumber:   int64(block.Number),
		BlockHash:     block.Hash.Hex(),
		ParentHash:    block.ParentHash.Hex(),
		BlockTime:     time.Unix(int64(block.Time), 0),
		IsOrphaned:    false,
		ScannedAt:     time.Now(),
	}

	err := p.blockRepo.Create(dbBlock)
	if err != nil {
		return fmt.Errorf("failed to save block %d: %w", block.Number, err)
	}

	p.logger.Debug("block saved", "chain_id", p.chainID, "block_number", block.Number, "block_hash", block.Hash.Hex())
	return nil
}

// DetectReorg detects if a reorg has occurred by checking block hash continuity
func (p *persistence) DetectReorg(ctx context.Context, blockNumber int64, currentBlockHash string) (bool, error) {
	if blockNumber <= 0 {
		return false, nil
	}

	// Get the previous block
	_, err := p.blockRepo.GetByChainIDAndBlockNumber(p.chainID, blockNumber-1)
	if err != nil {
		// If we can't find the previous block, assume no reorg
		return false, nil
	}

	// If the parent hash of current block doesn't match the stored hash, reorg occurred
	currentBlock, err := p.blockRepo.GetByChainIDAndBlockNumber(p.chainID, blockNumber)
	if err != nil {
		return false, nil
	}

	if currentBlock.BlockHash != currentBlockHash && currentBlock.BlockHash != "" {
		p.logger.Warn("reorg detected",
			"chain_id", p.chainID,
			"block_number", blockNumber,
			"expected_hash", currentBlockHash,
			"stored_hash", currentBlock.BlockHash,
		)
		return true, nil
	}

	return false, nil
}

// MarkBlocksOrphaned marks all blocks from the given block number as orphaned
func (p *persistence) MarkBlocksOrphaned(ctx context.Context, blockNumber int64) error {
	err := p.blockRepo.MarkOrphaned(p.chainID, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to mark blocks as orphaned from block %d: %w", blockNumber, err)
	}
	p.logger.Info("blocks marked as orphaned", "chain_id", p.chainID, "from_block", blockNumber)
	return nil
}
