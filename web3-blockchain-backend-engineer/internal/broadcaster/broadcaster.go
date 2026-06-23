package broadcaster

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
	"github.com/asset-platform/multi-chain-asset-platform/internal/withdrawalservice"
	"github.com/ethereum/go-ethereum/common"
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
	EstimateGas(ctx context.Context, req *EstimateGasRequest) (uint64, error)
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
	tokenRepo      repository.TokenRepository
	nonceRepo      NonceRepository
	rpcClient      RPCClient
	signer         Signer
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
	tokenRepo repository.TokenRepository,
	nonceRepo NonceRepository,
	rpcClient RPCClient,
	signer Signer,
	logger *slog.Logger,
) *Broadcaster {
	return &Broadcaster{
		db:             db,
		natsClient:     natsClient,
		withdrawalRepo: withdrawalRepo,
		tokenRepo:      tokenRepo,
		nonceRepo:      nonceRepo,
		rpcClient:      rpcClient,
		signer:         signer,
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
	if b.signer == nil {
		return "", fmt.Errorf("signer is not configured")
	}
	if b.rpcClient == nil {
		return "", fmt.Errorf("rpc client is not configured")
	}
	if !common.IsHexAddress(withdrawal.ToAddress) {
		return "", fmt.Errorf("invalid withdrawal destination address: %s", withdrawal.ToAddress)
	}

	token, err := b.resolveToken(withdrawal)
	if err != nil {
		return "", err
	}

	nonce, err := b.resolveNonce(ctx, withdrawal)
	if err != nil {
		return "", err
	}

	gasPrice, err := b.rpcClient.GasPrice(ctx, withdrawal.ChainID)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	amount, ok := new(big.Int).SetString(strings.TrimSpace(withdrawal.Amount), 10)
	if !ok || amount.Sign() <= 0 {
		return "", fmt.Errorf("invalid withdrawal amount: %s", withdrawal.Amount)
	}

	req, err := b.buildSignRequest(ctx, withdrawal, token, nonce, amount, gasPrice)
	if err != nil {
		return "", err
	}

	signedTx, err := b.signer.SignWithdrawal(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to sign withdrawal transaction: %w", err)
	}

	txHash, err := b.rpcClient.SendRawTransaction(ctx, withdrawal.ChainID, signedTx)
	if err != nil {
		if releaseErr := b.nonceRepo.Release(ctx, withdrawal.ChainID, withdrawal.FromAddress, int64(nonce)); releaseErr != nil {
			b.logger.Warn("failed to release nonce after send failure",
				"withdrawal_id", withdrawal.ID,
				"nonce", nonce,
				"error", releaseErr,
			)
		}
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	withdrawal.Nonce = int64(nonce)
	return txHash, nil
}

func (b *Broadcaster) resolveToken(withdrawal *repository.Withdrawal) (*repository.Token, error) {
	if b.tokenRepo == nil {
		return nil, fmt.Errorf("token repository is not configured")
	}

	token, err := b.tokenRepo.GetByID(withdrawal.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to load token metadata: %w", err)
	}
	return token, nil
}

func (b *Broadcaster) resolveNonce(ctx context.Context, withdrawal *repository.Withdrawal) (uint64, error) {
	if withdrawal.Nonce > 0 {
		return uint64(withdrawal.Nonce), nil
	}
	if b.nonceRepo == nil {
		return 0, fmt.Errorf("nonce repository is not configured")
	}

	chainNonce, err := b.rpcClient.NonceAt(ctx, withdrawal.ChainID, withdrawal.FromAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending nonce: %w", err)
	}
	allocatedNonce, err := b.nonceRepo.Allocate(ctx, withdrawal.ChainID, withdrawal.FromAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to allocate nonce: %w", err)
	}
	if uint64(allocatedNonce) < chainNonce {
		return chainNonce, nil
	}
	return uint64(allocatedNonce), nil
}

func (b *Broadcaster) buildSignRequest(
	ctx context.Context,
	withdrawal *repository.Withdrawal,
	token *repository.Token,
	nonce uint64,
	amount *big.Int,
	gasPrice *big.Int,
) (*SignRequest, error) {
	fromAddress := common.HexToAddress(withdrawal.FromAddress)
	toAddress := common.HexToAddress(withdrawal.ToAddress)

	req := &SignRequest{
		ChainID:  withdrawal.ChainID,
		Nonce:    nonce,
		To:       toAddress,
		Value:    amount,
		GasPrice: gasPrice,
		Token:    token,
	}

	gasLimit, err := b.estimateGasLimit(ctx, fromAddress, req)
	if err != nil {
		return nil, err
	}
	req.GasLimit = gasLimit

	return req, nil
}

func (b *Broadcaster) estimateGasLimit(ctx context.Context, from common.Address, req *SignRequest) (uint64, error) {
	callReq := &EstimateGasRequest{
		ChainID:  req.ChainID,
		From:     from,
		Value:    big.NewInt(0).Set(req.Value),
		GasPrice: req.GasPrice,
	}

	if req.Token == nil || req.Token.IsNative {
		callReq.To = ptrAddress(req.To)
		callReq.Value = big.NewInt(0).Set(req.Value)
	} else {
		signer, ok := b.signer.(*LocalHexKeySigner)
		if !ok {
			return 0, fmt.Errorf("unsupported signer implementation for erc20 gas estimation")
		}
		input, err := signer.erc20ABI.Pack("transfer", req.To, req.Value)
		if err != nil {
			return 0, fmt.Errorf("failed to encode erc20 calldata: %w", err)
		}
		contract := common.HexToAddress(req.Token.ContractAddress)
		callReq.To = ptrAddress(contract)
		callReq.Value = big.NewInt(0)
		callReq.Data = input
	}

	gas, err := b.rpcClient.EstimateGas(ctx, callReq)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	buffered := gas + gas/5
	if buffered < 21_000 {
		buffered = 21_000
	}
	return buffered, nil
}

func NewBroadcasterFromEnv(
	db *sql.DB,
	natsClient *nats.Conn,
	withdrawalRepo repository.WithdrawalRepository,
	tokenRepo repository.TokenRepository,
	nonceRepo NonceRepository,
	chainRepo repository.ChainRepository,
	providerRepo repository.RPCProviderRepository,
	logger *slog.Logger,
) (*Broadcaster, error) {
	privateKey := strings.TrimSpace(os.Getenv("WITHDRAWAL_SIGNER_PRIVATE_KEY"))
	if privateKey == "" {
		return nil, fmt.Errorf("WITHDRAWAL_SIGNER_PRIVATE_KEY is required")
	}

	signer, err := NewLocalHexKeySigner(privateKey)
	if err != nil {
		return nil, err
	}

	rpcClient := NewEVMRPCAdapter(chainRepo, providerRepo)
	return NewBroadcaster(
		db,
		natsClient,
		withdrawalRepo,
		tokenRepo,
		nonceRepo,
		rpcClient,
		signer,
		logger,
	), nil
}

func IsReceiptPending(err error) bool {
	return err == rpcgateway.ErrReceiptNotFound
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
