package reorgdetector

import (
	"context"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

// compensateDeposits marks deposits affected by reorg as orphaned
func (d *ReorgDetector) compensateDeposits(ctx context.Context, orphanedBlocks []*repository.Block) error {
	totalCompensated := 0

	for _, block := range orphanedBlocks {
		// Query deposits associated with this block number
		deposits, err := d.depositRepo.ListByBlockNumber(d.chainID, block.BlockNumber)
		if err != nil {
			d.logger.Warn("failed to list deposits for block",
				"block_number", block.BlockNumber,
				"error", err,
			)
			continue
		}

		for _, deposit := range deposits {
			// Only process confirmed deposits
			if deposit.Status != repository.DepositStatusConfirmed {
				continue
			}

			// Mark deposit as orphaned
			err := d.depositRepo.UpdateStatus(deposit.ID, repository.DepositStatusOrphaned)
			if err != nil {
				d.logger.Error("failed to mark deposit as orphaned",
					"deposit_id", deposit.ID,
					"tx_hash", deposit.TxHash,
					"error", err,
				)
				continue
			}

			d.logger.Info("deposit marked as orphaned",
				"deposit_id", deposit.ID,
				"chain_id", deposit.ChainID,
				"tx_hash", deposit.TxHash,
				"block_number", deposit.BlockNumber,
			)
			totalCompensated++
		}
	}

	d.logger.Info("deposit compensation completed",
		"chain_id", d.chainID,
		"total_orphaned", totalCompensated,
	)

	return nil
}

// compensateLedger creates reversal entries for ledger changes affected by reorg
func (d *ReorgDetector) compensateLedger(ctx context.Context, orphanedBlocks []*repository.Block) error {
	totalReversed := 0

	for _, block := range orphanedBlocks {
		// Query ledger entries associated with this block number
		entries, err := d.ledgerRepo.ListByBlockNumber(d.chainID, block.BlockNumber)
		if err != nil {
			d.logger.Warn("failed to list ledger entries for block",
				"block_number", block.BlockNumber,
				"error", err,
			)
			continue
		}

		for _, entry := range entries {
			// Skip entries that are already reversals
			if entry.EntryType == repository.LedgerEntryTypeReversal {
				continue
			}

			// Create a reversal entry with opposite direction
			reversal := &repository.LedgerEntry{
				AccountAddress: entry.AccountAddress,
				ChainID:        entry.ChainID,
				TokenID:        entry.TokenID,
				Direction:      d.oppositeDirection(entry.Direction),
				Amount:          entry.Amount,
				EntryType:       repository.LedgerEntryTypeReversal,
				ReferenceType:   repository.ReferenceTypeReorg,
				ReferenceID:    entry.ID,
				ReversalOf:     entry.ID,
				CreatedAt:      time.Now(),
			}

			err := d.ledgerRepo.Create(reversal)
			if err != nil {
				d.logger.Error("failed to create reversal ledger entry",
					"original_entry_id", entry.ID,
					"error", err,
				)
				continue
			}

			d.logger.Info("reversal ledger entry created",
				"original_entry_id", entry.ID,
				"reversal_entry_id", reversal.ID,
				"account_address", entry.AccountAddress,
				"chain_id", entry.ChainID,
				"amount", entry.Amount,
				"original_direction", entry.Direction,
				"reversal_direction", reversal.Direction,
			)
			totalReversed++
		}
	}

	d.logger.Info("ledger compensation completed",
		"chain_id", d.chainID,
		"total_reversed", totalReversed,
	)

	return nil
}

// isEntryRelatedToBlock checks if a ledger entry is related to a specific block
// This is a helper method that can be customized based on how block references are tracked
func (d *ReorgDetector) isEntryRelatedToBlock(entry *repository.LedgerEntry, block *repository.Block) bool {
	// Check if the entry's reference_id corresponds to a deposit or withdrawal in this block
	// This is a simplified implementation - in production you might have explicit block tracking
	// For now, we check if the entry was created around the same time as the block
	blockTime := block.ScannedAt
	entryTime := entry.CreatedAt

	// Consider entries within a 5-minute window of the block as potentially related
	timeDiff := entryTime.Sub(blockTime)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	return timeDiff < 5*time.Minute && entry.ChainID == block.ChainID
}

// oppositeDirection returns the opposite ledger direction for reversals
func (d *ReorgDetector) oppositeDirection(dir repository.LedgerDirection) repository.LedgerDirection {
	switch dir {
	case repository.LedgerDirectionCredit:
		return repository.LedgerDirectionDebit
	case repository.LedgerDirectionDebit:
		return repository.LedgerDirectionCredit
	default:
		d.logger.Warn("unknown ledger direction, returning original", "direction", dir)
		return dir
	}
}
