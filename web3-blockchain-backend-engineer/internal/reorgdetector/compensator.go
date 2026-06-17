package reorgdetector

import (
	"context"
	"encoding/json"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// handleReorg processes a detected chain reorganization
func (d *ReorgDetector) handleReorg(ctx context.Context, orphanedBlocks []*repository.Block, newBlockHash string) error {
	if len(orphanedBlocks) == 0 {
		return nil
	}

	firstOrphaned := orphanedBlocks[0]
	d.logger.Warn("processing reorg event",
		"chain_id", d.chainID,
		"fork_block", firstOrphaned.BlockNumber,
		"orphaned_count", len(orphanedBlocks),
		"new_block_hash", newBlockHash,
	)

	// 1. Mark all orphaned blocks in database
	if err := d.markBlocksOrphaned(ctx, orphanedBlocks); err != nil {
		d.logger.Error("failed to mark blocks as orphaned", "error", err)
		// Continue with other compensation steps even if this fails
	}

	// 2. Create and publish reorg event to NATS
	reorgEvent := &ReorgEvent{
		ChainID:      d.chainID,
		ForkBlock:    uint64(firstOrphaned.BlockNumber),
		NewBlockHash: newBlockHash,
		OldBlockHash: firstOrphaned.BlockHash,
	}

	if err := d.publishReorgEvent(reorgEvent); err != nil {
		d.logger.Error("failed to publish reorg event", "error", err)
		// Continue with other compensation steps even if this fails
	}

	// 3. Compensate deposits affected by this reorg
	if err := d.compensateDeposits(ctx, orphanedBlocks); err != nil {
		d.logger.Error("failed to compensate deposits", "error", err)
		// Continue with ledger compensation even if this fails
	}

	// 4. Compensate ledger entries affected by this reorg
	if err := d.compensateLedger(ctx, orphanedBlocks); err != nil {
		d.logger.Error("failed to compensate ledger entries", "error", err)
	}

	d.logger.Info("reorg compensation completed",
		"chain_id", d.chainID,
		"fork_block", firstOrphaned.BlockNumber,
		"orphaned_count", len(orphanedBlocks),
	)

	return nil
}

// markBlocksOrphaned updates the database to mark affected blocks as orphaned
func (d *ReorgDetector) markBlocksOrphaned(ctx context.Context, orphanedBlocks []*repository.Block) error {
	for _, block := range orphanedBlocks {
		err := d.blockRepo.MarkOrphaned(d.chainID, block.BlockNumber)
		if err != nil {
			d.logger.Error("failed to mark block orphaned",
				"chain_id", d.chainID,
				"block_number", block.BlockNumber,
				"error", err,
			)
			// Continue marking other blocks even if one fails
		} else {
			d.logger.Info("block marked as orphaned",
				"chain_id", d.chainID,
				"block_number", block.BlockNumber,
				"block_hash", block.BlockHash,
			)
		}
	}
	return nil
}

// publishReorgEvent sends the reorg event to NATS for other services to consume
func (d *ReorgDetector) publishReorgEvent(event *ReorgEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish to the reorg_events subject
	if err := d.natsClient.Publish("reorg_events", data); err != nil {
		return err
	}

	d.logger.Info("reorg event published to NATS",
		"subject", "reorg_events",
		"chain_id", event.ChainID,
		"fork_block", event.ForkBlock,
	)

	return nil
}

// SubscribeToReorgEvents creates a NATS subscription for reorg events
// This allows other services to be notified when reorgs occur
func (d *ReorgDetector) SubscribeToReorgEvents(handler func(*ReorgEvent)) (*nats.Subscription, error) {
	sub, err := d.natsClient.Subscribe("reorg_events", func(msg *nats.Msg) {
		var event ReorgEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			d.logger.Error("failed to unmarshal reorg event", "error", err)
			return
		}
		handler(&event)
	})
	if err != nil {
		return nil, err
	}

	d.logger.Info("subscribed to reorg events")
	return sub, nil
}
