package confirmworker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// NATS subjects
const (
	SubjectParsedEvents   = "parsed_events"
	SubjectDepositConfirm = "deposit_confirmed"
)

// ConfirmWorker processes deposit confirmations
type ConfirmWorker struct {
	natsClient    *nats.Conn
	depositRepo   repository.DepositRepository
	chainRepo     repository.ChainRepository
	rpcClient     RPCClient
	logger        *slog.Logger
	checkInterval time.Duration

	// Track deposits we're monitoring
	mu      sync.RWMutex
	tracked map[int64]*trackedDeposit // depositID -> trackedDeposit
}

// trackedDeposit holds state for a deposit being tracked
type trackedDeposit struct {
	Deposit        *repository.Deposit
	RequiredConfs  int
	LastKnownBlock uint64
	PublishFailed  bool // true if confirm succeeded but event publish failed
}

// RPCClient interface for blockchain RPC calls
type RPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

// NewConfirmWorker creates a new ConfirmWorker
func NewConfirmWorker(
	natsClient *nats.Conn,
	depositRepo repository.DepositRepository,
	chainRepo repository.ChainRepository,
	rpcClient RPCClient,
	logger *slog.Logger,
) *ConfirmWorker {
	return &ConfirmWorker{
		natsClient:    natsClient,
		depositRepo:   depositRepo,
		chainRepo:     chainRepo,
		rpcClient:     rpcClient,
		logger:        logger,
		checkInterval: 10 * time.Second,
		tracked:       make(map[int64]*trackedDeposit),
	}
}

// Start begins the confirmation worker loop
func (w *ConfirmWorker) Start(ctx context.Context) error {
	// Subscribe to parsed events
	sub, err := w.natsClient.Subscribe(SubjectParsedEvents, w.handleParsedEvent)
	if err != nil {
		w.logger.Error("Failed to subscribe to parsed_events", "error", err)
		return err
	}

	w.logger.Info("Confirm worker started, listening for parsed events")

	// Start tracking existing pending deposits
	if err := w.trackPendingDeposits(ctx); err != nil {
		w.logger.Warn("Failed to track pending deposits", "error", err)
	}

	// Start confirmation check loop
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			w.logger.Info("Confirm worker stopped")
			return ctx.Err()
		case <-ticker.C:
			w.checkConfirmations(ctx)
		}
	}
}

// handleParsedEvent processes a new parsed event
func (w *ConfirmWorker) handleParsedEvent(msg *nats.Msg) {
	var event ParsedEventMessage
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		w.logger.Error("Failed to unmarshal parsed event", "error", err)
		return
	}

	// Only process deposit-bearing events.
	if event.EventName != "Transfer" && event.EventName != "NativeTransfer" {
		return
	}

	// Look up the deposit by idempotency key
	idempotencyKey := MakeIdempotencyKey(event.ChainID, event.TxHash, event.LogIndex)
	deposit, err := w.depositRepo.GetByIdempotencyKey(idempotencyKey)
	if err != nil {
		w.logger.Debug("Deposit not found for event", "idempotency_key", idempotencyKey)
		return
	}

	// Only track deposits that are in detected or pending_confirmation status
	if deposit.Status != repository.DepositStatusDetected &&
		deposit.Status != repository.DepositStatusPendingConfirmation {
		return
	}

	// Get chain config for finality requirements
	chain, err := w.chainRepo.GetByChainID(event.ChainID)
	if err != nil {
		w.logger.Error("Failed to get chain config",
			"chain_id", event.ChainID,
			"error", err,
		)
		return
	}

	// Track this deposit
	w.trackDeposit(deposit, chain.FinalityConfirmations)

	msg.Ack()
}

// trackPendingDeposits loads and tracks deposits still awaiting confirmation
func (w *ConfirmWorker) trackPendingDeposits(ctx context.Context) error {
	deposits, err := w.depositRepo.ListByStatus(repository.DepositStatusDetected, 1000)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		chain, err := w.chainRepo.GetByChainID(deposit.ChainID)
		if err != nil {
			w.logger.Warn("Failed to get chain for deposit",
				"deposit_id", deposit.ID,
				"chain_id", deposit.ChainID,
			)
			continue
		}
		w.trackDeposit(deposit, chain.FinalityConfirmations)
	}

	// Also track pending_confirmation deposits
	pending, err := w.depositRepo.ListByStatus(repository.DepositStatusPendingConfirmation, 1000)
	if err != nil {
		return err
	}

	for _, deposit := range pending {
		chain, err := w.chainRepo.GetByChainID(deposit.ChainID)
		if err != nil {
			continue
		}
		w.trackDeposit(deposit, chain.FinalityConfirmations)
	}

	// Confirmed deposits whose event was not durably marked as published must be retried after restart.
	confirmedUnpublished, err := w.depositRepo.ListConfirmedUnpublished(1000)
	if err != nil {
		return err
	}

	for _, deposit := range confirmedUnpublished {
		chain, err := w.chainRepo.GetByChainID(deposit.ChainID)
		if err != nil {
			continue
		}
		w.trackDepositWithPublishFlag(deposit, chain.FinalityConfirmations, true)
	}

	w.logger.Info("Tracking deposits",
		"detected", len(deposits),
		"pending_confirmation", len(pending),
		"confirmed_unpublished", len(confirmedUnpublished),
	)

	return nil
}

// trackDeposit adds a deposit to the tracking map
func (w *ConfirmWorker) trackDeposit(deposit *repository.Deposit, requiredConfs int) {
	w.trackDepositWithPublishFlag(deposit, requiredConfs, false)
}

// trackDepositWithPublishFlag adds a deposit to the tracking map with a publish flag
func (w *ConfirmWorker) trackDepositWithPublishFlag(deposit *repository.Deposit, requiredConfs int, publishFailed bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Skip if already confirmed or orphaned (unless publish previously failed)
	if deposit.Status == repository.DepositStatusOrphaned {
		return
	}
	if deposit.Status == repository.DepositStatusConfirmed && !publishFailed {
		return
	}

	w.tracked[deposit.ID] = &trackedDeposit{
		Deposit:       deposit,
		RequiredConfs: requiredConfs,
		PublishFailed: publishFailed,
	}
}

// checkConfirmations periodically checks and updates confirmation counts
func (w *ConfirmWorker) checkConfirmations(ctx context.Context) {
	w.mu.RLock()
	tracked := make([]*trackedDeposit, 0, len(w.tracked))
	for _, td := range w.tracked {
		tracked = append(tracked, td)
	}
	w.mu.RUnlock()

	if len(tracked) == 0 {
		return
	}

	// Get current block number
	currentBlock, err := w.rpcClient.BlockNumber(ctx)
	if err != nil {
		w.logger.Warn("Failed to get current block number", "error", err)
		return
	}

	for _, td := range tracked {
		w.processDeposit(ctx, td, currentBlock)
	}
}

// processDeposit checks and updates confirmation count for a single deposit
func (w *ConfirmWorker) processDeposit(ctx context.Context, td *trackedDeposit, currentBlock uint64) {
	deposit := td.Deposit

	// If publish previously failed, retry publishing the confirmed event
	if td.PublishFailed && deposit.Status == repository.DepositStatusConfirmed {
		if err := w.publishDepositConfirmed(deposit); err != nil {
			w.logger.Warn("Retry publish failed, will retry next cycle",
				"deposit_id", deposit.ID,
				"error", err,
			)
			return
		}
		if err := w.depositRepo.MarkConfirmedEventPublished(deposit.ID); err != nil {
			w.logger.Warn("Retry publish succeeded but failed to mark published",
				"deposit_id", deposit.ID,
				"error", err,
			)
			return
		}
		// Publish succeeded, remove from tracking
		w.mu.Lock()
		delete(w.tracked, deposit.ID)
		w.mu.Unlock()
		w.logger.Info("Retry publish succeeded",
			"deposit_id", deposit.ID,
		)
		return
	}

	// Calculate current confirmations
	var confirmations int
	if currentBlock > uint64(deposit.BlockNumber) {
		confirmations = int(currentBlock - uint64(deposit.BlockNumber))
	}

	// Update confirmation count in database with absolute value (P0 fix)
	if confirmations > deposit.Confirmations {
		if err := w.depositRepo.SetConfirmations(deposit.ID, confirmations); err != nil {
			w.logger.Warn("Failed to set confirmations",
				"deposit_id", deposit.ID,
				"confirmations", confirmations,
				"error", err,
			)
		}
		deposit.Confirmations = confirmations
	}

	// State transitions
	switch deposit.Status {
	case repository.DepositStatusDetected:
		// Transition to pending_confirmation after first confirmation
		if confirmations >= 1 {
			if err := w.depositRepo.UpdateStatus(deposit.ID, repository.DepositStatusPendingConfirmation); err != nil {
				w.logger.Warn("Failed to update deposit status",
					"deposit_id", deposit.ID,
					"error", err,
				)
			} else {
				deposit.Status = repository.DepositStatusPendingConfirmation
				w.logger.Info("Deposit transitioned to pending_confirmation",
					"deposit_id", deposit.ID,
					"confirmations", confirmations,
				)
			}
		}

	case repository.DepositStatusPendingConfirmation:
		// Use atomic conditional update to avoid race conditions (P0 fix)
		// Only confirms if: status is still pending_confirmation AND confirmations >= target
		updated, err := w.depositRepo.ConfirmWithCondition(deposit.ID, confirmations, td.RequiredConfs)
		if err != nil {
			w.logger.Warn("Failed to confirm deposit",
				"deposit_id", deposit.ID,
				"error", err,
			)
			return
		}

		if updated {
			deposit.Status = repository.DepositStatusConfirmed
			deposit.Confirmations = confirmations
			w.logger.Info("Deposit confirmed via conditional update",
				"deposit_id", deposit.ID,
				"chain_id", deposit.ChainID,
				"tx_hash", deposit.TxHash,
				"confirmations", confirmations,
			)

			// Publish confirmed event - only remove from tracking if publish succeeds (P0 fix)
			if err := w.publishDepositConfirmed(deposit); err != nil {
				w.logger.Error("Deposit confirmed but event publish failed - will retry",
					"deposit_id", deposit.ID,
					"error", err,
				)
				// Mark publish as failed so we retry on next cycle
				td.PublishFailed = true
				return
			}
			if err := w.depositRepo.MarkConfirmedEventPublished(deposit.ID); err != nil {
				w.logger.Error("Deposit confirmed event published but failed to mark published - will retry",
					"deposit_id", deposit.ID,
					"error", err,
				)
				td.PublishFailed = true
				return
			}
			deposit.ConfirmedEventPublished = true

			// Stop tracking only after successful publish
			w.mu.Lock()
			delete(w.tracked, deposit.ID)
			w.mu.Unlock()
		}
	}
}

// publishDepositConfirmed publishes a confirmed deposit event to NATS
// Returns error if publish fails so caller can retry
func (w *ConfirmWorker) publishDepositConfirmed(deposit *repository.Deposit) error {
	event := DepositConfirmedEvent{
		DepositID:   deposit.ID,
		ChainID:     deposit.ChainID,
		TokenID:     deposit.TokenID,
		Account:     deposit.ToAddress,
		Amount:      deposit.Amount,
		TxHash:      deposit.TxHash,
		BlockNumber: deposit.BlockNumber,
	}

	data, err := json.Marshal(event)
	if err != nil {
		w.logger.Error("Failed to marshal deposit confirmed event", "error", err)
		return err
	}

	if err := w.natsClient.Publish(SubjectDepositConfirm, data); err != nil {
		w.logger.Error("Failed to publish deposit confirmed event", "error", err)
		return err
	}

	w.logger.Info("Published deposit_confirmed event",
		"deposit_id", deposit.ID,
		"subject", SubjectDepositConfirm,
	)
	return nil
}

// ParsedEventMessage represents a parsed event from NATS
type ParsedEventMessage struct {
	ChainID      int64  `json:"chain_id"`
	TokenID      int64  `json:"token_id"`
	TokenAddress string `json:"token_address"`
	TxHash       string `json:"tx_hash"`
	LogIndex     uint   `json:"log_index"`
	BlockNumber  int64  `json:"block_number"`
	BlockHash    string `json:"block_hash"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
	EventName    string `json:"event_name"`
}

// DepositConfirmedEvent represents a confirmed deposit event
type DepositConfirmedEvent struct {
	DepositID   int64  `json:"deposit_id"`
	ChainID     int64  `json:"chain_id"`
	TokenID     int64  `json:"token_id"`
	Account     string `json:"account"`
	Amount      string `json:"amount"`
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
}

// MakeIdempotencyKey creates an idempotency key for a deposit
func MakeIdempotencyKey(chainID int64, txHash string, logIndex uint) string {
	return fmt.Sprintf("%d:%s:%d", chainID, txHash, logIndex)
}
