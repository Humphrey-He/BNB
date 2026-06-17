package withdrawalservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// NATS subjects for withdrawal events
const (
	SubjectWithdrawalCreated   = "withdrawal_created"
	SubjectWithdrawalApproved  = "withdrawal_approved"
	SubjectWithdrawalBroadcast = "withdrawal_broadcast"
)

// WithdrawalWorker processes withdrawal requests through their lifecycle:
// created → risk_checking → approved → signing → broadcasting → broadcasted → confirmed/failed
type WithdrawalWorker struct {
	natsClient     *nats.Conn
	withdrawalRepo repository.WithdrawalRepository
	balanceRepo    repository.BalanceRepository
	logger         *slog.Logger
	checkInterval  time.Duration
	riskConfig     RiskConfig
}

// NewWithdrawalWorker creates a new WithdrawalWorker
func NewWithdrawalWorker(
	natsClient *nats.Conn,
	withdrawalRepo repository.WithdrawalRepository,
	balanceRepo repository.BalanceRepository,
	logger *slog.Logger,
) *WithdrawalWorker {
	return &WithdrawalWorker{
		natsClient:     natsClient,
		withdrawalRepo: withdrawalRepo,
		balanceRepo:    balanceRepo,
		logger:         logger,
		checkInterval:  5 * time.Second,
		riskConfig:     LoadRiskConfig(),
	}
}

// Start begins the withdrawal worker loop
func (w *WithdrawalWorker) Start(ctx context.Context) error {
	// Subscribe to new withdrawal events
	sub, err := w.natsClient.Subscribe(SubjectWithdrawalCreated, w.handleWithdrawalCreated)
	if err != nil {
		w.logger.Error("Failed to subscribe to withdrawal_created", "error", err)
		return err
	}

	w.logger.Info("Withdrawal worker started, listening for withdrawal events")

	// Start processing loop
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			w.logger.Info("Withdrawal worker stopped")
			return ctx.Err()
		case <-ticker.C:
			w.processPendingWithdrawals(ctx)
		}
	}
}

// handleWithdrawalCreated processes a new withdrawal request
func (w *WithdrawalWorker) handleWithdrawalCreated(msg *nats.Msg) {
	var event WithdrawalCreatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		w.logger.Error("Failed to unmarshal withdrawal created event", "error", err)
		return
	}

	w.logger.Info("Processing new withdrawal",
		"withdrawal_id", event.WithdrawalID,
		"to_address", event.ToAddress,
		"amount", event.Amount,
	)

	// Process the withdrawal through risk check
	if err := w.processWithdrawalRiskCheck(context.Background(), event.WithdrawalID); err != nil {
		w.logger.Error("Failed to process withdrawal risk check",
			"withdrawal_id", event.WithdrawalID,
			"error", err,
		)
		return
	}

	msg.Ack()
}

// processPendingWithdrawals processes withdrawals in each stage
func (w *WithdrawalWorker) processPendingWithdrawals(ctx context.Context) {
	// Process withdrawals in risk_checking status
	if err := w.processRiskCheckStage(ctx); err != nil {
		w.logger.Warn("Failed to process risk check stage", "error", err)
	}

	// Process withdrawals in approved status (ready for signing)
	if err := w.processApprovedStage(ctx); err != nil {
		w.logger.Warn("Failed to process approved stage", "error", err)
	}
}

// processRiskCheckStage moves withdrawals from created → risk_checking → approved/failed
func (w *WithdrawalWorker) processRiskCheckStage(ctx context.Context) error {
	// Get withdrawals in 'created' status
	withdrawals, err := w.withdrawalRepo.ListByStatus(repository.WithdrawalStatusCreated, 100)
	if err != nil {
		return fmt.Errorf("failed to list created withdrawals: %w", err)
	}

	for _, withdrawal := range withdrawals {
		// Update to risk_checking
		if err := w.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusRiskChecking); err != nil {
			w.logger.Warn("Failed to update withdrawal status to risk_checking",
				"withdrawal_id", withdrawal.ID,
				"error", err,
			)
			continue
		}

		decision := w.evaluateRisk(withdrawal)
		if decision.Status != repository.WithdrawalStatusApproved {
			w.handleRiskDecision(withdrawal, decision, "scheduled-risk-check")
			continue
		}

		// Risk check passed - approve
		if err := w.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusApproved); err != nil {
			w.logger.Warn("Failed to update withdrawal status to approved",
				"withdrawal_id", withdrawal.ID,
				"error", err,
			)
			continue
		}

		w.logger.Info("Withdrawal approved",
			"withdrawal_id", withdrawal.ID,
			"amount", withdrawal.Amount,
		)

		// Publish approved event for broadcaster
		w.publishWithdrawalApproved(withdrawal)
	}

	return nil
}

// performRiskCheck performs risk checks on a withdrawal
// Returns error if check fails
func (w *WithdrawalWorker) performRiskCheck(ctx context.Context, withdrawal *repository.Withdrawal) error {
	_ = ctx
	decision := w.evaluateRisk(withdrawal)
	if decision.Status != repository.WithdrawalStatusApproved {
		return fmt.Errorf("%s", decision.Reason)
	}

	w.logger.Debug("Risk check passed",
		"withdrawal_id", withdrawal.ID,
		"to_address", withdrawal.ToAddress,
	)

	return nil
}

// processApprovedStage moves approved withdrawals to signing/broadcasting
func (w *WithdrawalWorker) processApprovedStage(ctx context.Context) error {
	withdrawals, err := w.withdrawalRepo.ListByStatus(repository.WithdrawalStatusApproved, 100)
	if err != nil {
		return fmt.Errorf("failed to list approved withdrawals: %w", err)
	}

	for _, withdrawal := range withdrawals {
		// Move to signing status - broadcaster will pick this up
		if err := w.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusSigning); err != nil {
			w.logger.Warn("Failed to update withdrawal status to signing",
				"withdrawal_id", withdrawal.ID,
				"error", err,
			)
			continue
		}

		// Publish to broadcast subject
		w.publishWithdrawalBroadcast(withdrawal)
	}

	return nil
}

// processWithdrawalRiskCheck handles a new withdrawal from the event
func (w *WithdrawalWorker) processWithdrawalRiskCheck(ctx context.Context, withdrawalID int64) error {
	withdrawal, err := w.withdrawalRepo.GetByID(withdrawalID)
	if err != nil {
		return fmt.Errorf("failed to get withdrawal: %w", err)
	}

	// Update to risk_checking
	if err := w.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusRiskChecking); err != nil {
		return fmt.Errorf("failed to update status to risk_checking: %w", err)
	}

	decision := w.evaluateRisk(withdrawal)
	if decision.Status != repository.WithdrawalStatusApproved {
		w.handleRiskDecision(withdrawal, decision, "event-risk-check")
		return fmt.Errorf("%s", decision.Reason)
	}

	// Approve
	if err := w.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusApproved); err != nil {
		return fmt.Errorf("failed to update status to approved: %w", err)
	}

	// Publish approved event
	w.publishWithdrawalApproved(withdrawal)

	return nil
}

func (w *WithdrawalWorker) handleRiskDecision(withdrawal *repository.Withdrawal, decision riskDecision, source string) {
	withdrawal.FailureReason = decision.Reason
	withdrawal.Status = decision.Status
	if updateErr := w.withdrawalRepo.Update(withdrawal); updateErr != nil {
		w.logger.Warn("Failed to persist risk decision",
			"withdrawal_id", withdrawal.ID,
			"status", decision.Status,
			"error", updateErr,
		)
		return
	}

	if decision.Status == repository.WithdrawalStatusManualReview {
		w.logger.Warn("Withdrawal moved to manual review",
			"withdrawal_id", withdrawal.ID,
			"reason", decision.Reason,
			"source", source,
		)
		return
	}

	w.logger.Error("Withdrawal risk check failed",
		"withdrawal_id", withdrawal.ID,
		"reason", decision.Reason,
		"source", source,
	)
}

// publishWithdrawalApproved publishes an approved withdrawal event
func (w *WithdrawalWorker) publishWithdrawalApproved(withdrawal *repository.Withdrawal) {
	event := WithdrawalApprovedEvent{
		WithdrawalID: withdrawal.ID,
		ChainID:      withdrawal.ChainID,
		TokenID:      withdrawal.TokenID,
		FromAddress:  withdrawal.FromAddress,
		ToAddress:    withdrawal.ToAddress,
		Amount:       withdrawal.Amount,
	}

	data, err := json.Marshal(event)
	if err != nil {
		w.logger.Error("Failed to marshal withdrawal approved event", "error", err)
		return
	}

	if err := w.natsClient.Publish(SubjectWithdrawalApproved, data); err != nil {
		w.logger.Error("Failed to publish withdrawal approved event", "error", err)
		return
	}

	w.logger.Info("Published withdrawal_approved event",
		"withdrawal_id", withdrawal.ID,
	)
}

// publishWithdrawalBroadcast publishes a withdrawal ready for broadcast
func (w *WithdrawalWorker) publishWithdrawalBroadcast(withdrawal *repository.Withdrawal) {
	event := WithdrawalBroadcastEvent{
		WithdrawalID: withdrawal.ID,
		ChainID:      withdrawal.ChainID,
		TokenID:      withdrawal.TokenID,
		FromAddress:  withdrawal.FromAddress,
		ToAddress:    withdrawal.ToAddress,
		Amount:       withdrawal.Amount,
		Nonce:        withdrawal.Nonce,
	}

	data, err := json.Marshal(event)
	if err != nil {
		w.logger.Error("Failed to marshal withdrawal broadcast event", "error", err)
		return
	}

	if err := w.natsClient.Publish(SubjectWithdrawalBroadcast, data); err != nil {
		w.logger.Error("Failed to publish withdrawal broadcast event", "error", err)
		return
	}

	w.logger.Info("Published withdrawal_broadcast event",
		"withdrawal_id", withdrawal.ID,
	)
}

// isValidAddress performs basic address validation
func isValidAddress(addr string) bool {
	if len(addr) != 42 {
		return false
	}
	if addr[:2] != "0x" {
		return false
	}
	// Basic hex check
	for _, c := range addr[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// WithdrawalCreatedEvent represents a new withdrawal request event
type WithdrawalCreatedEvent struct {
	WithdrawalID   int64  `json:"withdrawal_id"`
	ChainID        int64  `json:"chain_id"`
	TokenID        int64  `json:"token_id"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Amount         string `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

// WithdrawalApprovedEvent represents an approved withdrawal ready for signing
type WithdrawalApprovedEvent struct {
	WithdrawalID int64  `json:"withdrawal_id"`
	ChainID      int64  `json:"chain_id"`
	TokenID      int64  `json:"token_id"`
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"`
}

// WithdrawalBroadcastEvent represents a withdrawal ready to be broadcast
type WithdrawalBroadcastEvent struct {
	WithdrawalID int64  `json:"withdrawal_id"`
	ChainID      int64  `json:"chain_id"`
	TokenID      int64  `json:"token_id"`
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"`
	Nonce        int64  `json:"nonce"`
}
