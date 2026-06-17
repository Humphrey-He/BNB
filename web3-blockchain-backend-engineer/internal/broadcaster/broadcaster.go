package broadcaster

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/withdrawalservice"
	"github.com/nats-io/nats.go"
)

// NATS subjects
const (
	SubjectWithdrawalBroadcast   = "withdrawal_broadcast"
	SubjectWithdrawalBroadcasted = "withdrawal_broadcasted"
)

// RPCClient interface for blockchain RPC calls
type RPCClient interface {
	GasPrice(ctx context.Context, chainID int64) (*big.Int, error)
	SendRawTransaction(ctx context.Context, chainID int64, signedTx []byte) (string, error)
	NonceAt(ctx context.Context, chainID int64, address string) (uint64, error)
	GetTransactionReceipt(ctx context.Context, chainID int64, txHash string) (*TxReceipt, error)
}

type TxReceipt struct {
	Status      uint64
	BlockNumber uint64
}

// Broadcaster handles transaction signing and broadcasting
type Broadcaster struct {
	db             *sql.DB
	natsClient     *nats.Conn
	withdrawalRepo repository.WithdrawalRepository
	nonceRepo      NonceRepository
	rpcClient      RPCClient
	logger         *slog.Logger
	checkInterval  time.Duration
	mu             sync.RWMutex
	tracked        map[int64]*trackedWithdrawal // withdrawalID -> trackedWithdrawal
}

// NonceRepository manages nonce allocation
type NonceRepository interface {
	Allocate(ctx context.Context, chainID int64, address string) (int64, error)
	Release(ctx context.Context, chainID int64, address string, nonce int64) error
}

// trackedWithdrawal holds state for a withdrawal being broadcast
type trackedWithdrawal struct {
	Withdrawal *repository.Withdrawal
	Nonce      int64
	SignedTx   []byte
	RetryCount int
	LastRetry  time.Time
}

// NewBroadcaster creates a new Broadcaster
func NewBroadcaster(
	db *sql.DB,
	natsClient *nats.Conn,
	withdrawalRepo repository.WithdrawalRepository,
	nonceRepo NonceRepository,
	rpcClient RPCClient,
	logger *slog.Logger,
) *Broadcaster {
	return &Broadcaster{
		db:             db,
		natsClient:     natsClient,
		withdrawalRepo: withdrawalRepo,
		nonceRepo:      nonceRepo,
		rpcClient:      rpcClient,
		logger:         logger,
		checkInterval:  3 * time.Second,
		tracked:        make(map[int64]*trackedWithdrawal),
	}
}

// Start begins the broadcaster loop
func (b *Broadcaster) Start(ctx context.Context) error {
	// Subscribe to withdrawal broadcast events
	sub, err := b.natsClient.Subscribe(SubjectWithdrawalBroadcast, b.handleWithdrawalBroadcast)
	if err != nil {
		b.logger.Error("Failed to subscribe to withdrawal_broadcast", "error", err)
		return err
	}

	b.logger.Info("Broadcaster started, listening for withdrawal broadcast events")

	// Start retry loop
	ticker := time.NewTicker(b.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			b.logger.Info("Broadcaster stopped")
			return ctx.Err()
		case <-ticker.C:
			b.retryPending(ctx)
		}
	}
}

// handleWithdrawalBroadcast processes a withdrawal ready for broadcast
func (b *Broadcaster) handleWithdrawalBroadcast(msg *nats.Msg) {
	var event withdrawalservice.WithdrawalBroadcastEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		b.logger.Error("Failed to unmarshal withdrawal broadcast event", "error", err)
		return
	}

	b.logger.Info("Processing withdrawal for broadcast",
		"withdrawal_id", event.WithdrawalID,
		"chain_id", event.ChainID,
	)

	// Get withdrawal details
	withdrawal, err := b.withdrawalRepo.GetByID(event.WithdrawalID)
	if err != nil {
		b.logger.Error("Failed to get withdrawal", "withdrawal_id", event.WithdrawalID, "error", err)
		return
	}

	// Update status to broadcasting
	if err := b.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusBroadcasting); err != nil {
		b.logger.Warn("Failed to update status to broadcasting", "error", err)
	}

	// Broadcast the transaction
	txHash, err := b.broadcastTransaction(context.Background(), withdrawal)
	if err != nil {
		b.logger.Error("Failed to broadcast transaction",
			"withdrawal_id", withdrawal.ID,
			"error", err,
		)
		// Track for retry
		b.trackWithdrawal(withdrawal, 0, nil)
		msg.Ack() // Ack to prevent redelivery, we'll handle retry via ticker
		return
	}

	// Success - update withdrawal with tx hash
	withdrawal.TxHash = txHash
	withdrawal.Status = repository.WithdrawalStatusBroadcasted
	if err := b.withdrawalRepo.Update(withdrawal); err != nil {
		b.logger.Warn("Failed to update withdrawal with tx hash", "error", err)
	}

	if b.db != nil {
		if err := settleBroadcastedWithdrawal(context.Background(), b.db, withdrawal); err != nil {
			b.logger.Error("Failed to settle broadcasted withdrawal",
				"withdrawal_id", withdrawal.ID,
				"error", err,
			)
			return
		}
	}

	// Publish broadcasted event
	b.publishWithdrawalBroadcasted(withdrawal)

	b.logger.Info("Withdrawal broadcasted successfully",
		"withdrawal_id", withdrawal.ID,
		"tx_hash", txHash,
	)

	msg.Ack()
}

// broadcastTransaction signs and broadcasts a withdrawal transaction
func (b *Broadcaster) broadcastTransaction(ctx context.Context, withdrawal *repository.Withdrawal) (string, error) {
	// Get nonce - use allocated nonce or fetch from chain
	var nonce int64
	if withdrawal.Nonce > 0 {
		nonce = withdrawal.Nonce
	} else {
		chainNonce, err := b.rpcClient.NonceAt(ctx, withdrawal.ChainID, withdrawal.FromAddress)
		if err != nil {
			return "", fmt.Errorf("failed to get nonce: %w", err)
		}
		nonce = int64(chainNonce)
	}

	// Get gas price
	gasPrice, err := b.rpcClient.GasPrice(ctx, withdrawal.ChainID)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Build unsigned transaction data
	// In production, this would build a proper signed transaction
	txData := buildTransactionData(withdrawal, nonce, gasPrice)

	// Sign the transaction (placeholder - in production uses proper key management)
	signedTx := b.signTransaction(txData, withdrawal.FromAddress)

	// Send to blockchain
	txHash, err := b.rpcClient.SendRawTransaction(ctx, withdrawal.ChainID, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return txHash, nil
}

// buildTransactionData creates the transaction data for a withdrawal
func buildTransactionData(withdrawal *repository.Withdrawal, nonce int64, gasPrice *big.Int) []byte {
	// In production, this would build an EIP-1559 or legacy transaction
	// For now, return a placeholder
	var data strings.Builder
	data.WriteString(fmt.Sprintf("chain=%d", withdrawal.ChainID))
	data.WriteString(fmt.Sprintf("&to=%s", withdrawal.ToAddress))
	data.WriteString(fmt.Sprintf("&value=%s", withdrawal.Amount))
	data.WriteString(fmt.Sprintf("&nonce=%d", nonce))
	data.WriteString(fmt.Sprintf("&gasPrice=%s", gasPrice.String()))
	return []byte(data.String())
}

// signTransaction signs transaction data
// In production, this would use proper key management (HSM, KMS, etc.)
func (b *Broadcaster) signTransaction(txData []byte, fromAddress string) []byte {
	// Placeholder signature - in production this would:
	// 1. Retrieve private key from secure key management
	// 2. Sign the transaction hash
	// 3. Return the RLP-encoded signed transaction
	b.logger.Debug("Signing transaction", "from", fromAddress)

	// Return txData as placeholder for signed bytes
	// Real implementation would return: rlp.Encode(signedTx)
	return append([]byte("signed:"), txData...)
}

// trackWithdrawal adds a withdrawal to the retry tracking
func (b *Broadcaster) trackWithdrawal(withdrawal *repository.Withdrawal, nonce int64, signedTx []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tracked[withdrawal.ID] = &trackedWithdrawal{
		Withdrawal: withdrawal,
		Nonce:      nonce,
		SignedTx:   signedTx,
		RetryCount: 0,
		LastRetry:  time.Now(),
	}
}

// retryPending retries failed broadcasts
func (b *Broadcaster) retryPending(ctx context.Context) {
	if err := b.reconcileBroadcasted(ctx); err != nil {
		b.logger.Warn("Failed to reconcile broadcasted withdrawals", "error", err)
	}

	b.mu.Lock()
	var toRetry []*trackedWithdrawal
	for _, tw := range b.tracked {
		// Retry every 30 seconds
		if time.Since(tw.LastRetry) > 30*time.Second && tw.RetryCount < 5 {
			toRetry = append(toRetry, tw)
		}
	}
	b.mu.Unlock()

	for _, tw := range toRetry {
		b.retryBroadcast(ctx, tw)
	}
}

func (b *Broadcaster) reconcileBroadcasted(ctx context.Context) error {
	withdrawals, err := b.withdrawalRepo.ListByStatus(repository.WithdrawalStatusBroadcasted, 100)
	if err != nil {
		return fmt.Errorf("failed to list broadcasted withdrawals: %w", err)
	}

	for _, withdrawal := range withdrawals {
		if strings.TrimSpace(withdrawal.TxHash) == "" {
			continue
		}

		receipt, err := b.rpcClient.GetTransactionReceipt(ctx, withdrawal.ChainID, withdrawal.TxHash)
		if err != nil {
			continue
		}

		switch receipt.Status {
		case 1:
			if err := b.withdrawalRepo.UpdateStatus(withdrawal.ID, repository.WithdrawalStatusConfirmed); err != nil {
				b.logger.Warn("Failed to confirm withdrawal from receipt",
					"withdrawal_id", withdrawal.ID,
					"tx_hash", withdrawal.TxHash,
					"error", err,
				)
				continue
			}
			b.logger.Info("Withdrawal confirmed from receipt",
				"withdrawal_id", withdrawal.ID,
				"tx_hash", withdrawal.TxHash,
				"block_number", receipt.BlockNumber,
			)
		case 0:
			withdrawal.Status = repository.WithdrawalStatusFailed
			withdrawal.FailureReason = "on-chain transaction reverted"
			if err := b.withdrawalRepo.Update(withdrawal); err != nil {
				b.logger.Warn("Failed to mark withdrawal failed from receipt",
					"withdrawal_id", withdrawal.ID,
					"tx_hash", withdrawal.TxHash,
					"error", err,
				)
				continue
			}
			if b.db != nil {
				if err := compensateRevertedWithdrawal(ctx, b.db, withdrawal); err != nil {
					b.logger.Error("Failed to compensate reverted withdrawal",
						"withdrawal_id", withdrawal.ID,
						"tx_hash", withdrawal.TxHash,
						"error", err,
					)
					continue
				}
			}
			b.logger.Warn("Withdrawal marked failed from receipt",
				"withdrawal_id", withdrawal.ID,
				"tx_hash", withdrawal.TxHash,
				"block_number", receipt.BlockNumber,
			)
		}
	}

	return nil
}

// retryBroadcast retries a failed broadcast
func (b *Broadcaster) retryBroadcast(ctx context.Context, tw *trackedWithdrawal) {
	b.mu.Lock()
	tw.RetryCount++
	tw.LastRetry = time.Now()
	b.mu.Unlock()

	txHash, err := b.broadcastTransaction(ctx, tw.Withdrawal)
	if err != nil {
		b.logger.Warn("Retry broadcast failed",
			"withdrawal_id", tw.Withdrawal.ID,
			"retry_count", tw.RetryCount,
			"error", err,
		)
		if tw.RetryCount >= 5 {
			// Max retries - mark as failed
			tw.Withdrawal.FailureReason = fmt.Sprintf("broadcast failed after %d retries: %s", tw.RetryCount, err.Error())
			tw.Withdrawal.Status = repository.WithdrawalStatusFailed
			if updateErr := b.withdrawalRepo.Update(tw.Withdrawal); updateErr != nil {
				b.logger.Warn("Failed to update withdrawal to failed", "error", updateErr)
			}
			if b.db != nil {
				if releaseErr := releaseFailedWithdrawal(ctx, b.db, tw.Withdrawal); releaseErr != nil {
					b.logger.Error("Failed to release failed withdrawal frozen balance",
						"withdrawal_id", tw.Withdrawal.ID,
						"error", releaseErr,
					)
				}
			}
			// Remove from tracking
			b.mu.Lock()
			delete(b.tracked, tw.Withdrawal.ID)
			b.mu.Unlock()
		}
		return
	}

	// Success
	tw.Withdrawal.TxHash = txHash
	tw.Withdrawal.Status = repository.WithdrawalStatusBroadcasted
	if err := b.withdrawalRepo.Update(tw.Withdrawal); err != nil {
		b.logger.Warn("Failed to update withdrawal with tx hash", "error", err)
	}

	if b.db != nil {
		if err := settleBroadcastedWithdrawal(ctx, b.db, tw.Withdrawal); err != nil {
			b.logger.Error("Failed to settle retry-broadcasted withdrawal",
				"withdrawal_id", tw.Withdrawal.ID,
				"error", err,
			)
			return
		}
	}

	b.publishWithdrawalBroadcasted(tw.Withdrawal)

	// Remove from tracking
	b.mu.Lock()
	delete(b.tracked, tw.Withdrawal.ID)
	b.mu.Unlock()

	b.logger.Info("Retry broadcast succeeded",
		"withdrawal_id", tw.Withdrawal.ID,
		"tx_hash", txHash,
	)
}

// publishWithdrawalBroadcasted publishes a successfully broadcasted event
func (b *Broadcaster) publishWithdrawalBroadcasted(withdrawal *repository.Withdrawal) {
	event := WithdrawalBroadcastedEvent{
		WithdrawalID: withdrawal.ID,
		ChainID:      withdrawal.ChainID,
		TokenID:      withdrawal.TokenID,
		FromAddress:  withdrawal.FromAddress,
		ToAddress:    withdrawal.ToAddress,
		Amount:       withdrawal.Amount,
		TxHash:       withdrawal.TxHash,
	}

	data, err := json.Marshal(event)
	if err != nil {
		b.logger.Error("Failed to marshal withdrawal broadcasted event", "error", err)
		return
	}

	if err := b.natsClient.Publish(SubjectWithdrawalBroadcasted, data); err != nil {
		b.logger.Error("Failed to publish withdrawal broadcasted event", "error", err)
		return
	}

	b.logger.Info("Published withdrawal_broadcasted event",
		"withdrawal_id", withdrawal.ID,
		"tx_hash", withdrawal.TxHash,
	)
}

// WithdrawalBroadcastedEvent represents a broadcasted withdrawal event
type WithdrawalBroadcastedEvent struct {
	WithdrawalID int64  `json:"withdrawal_id"`
	ChainID      int64  `json:"chain_id"`
	TokenID      int64  `json:"token_id"`
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"`
	TxHash       string `json:"tx_hash"`
}
