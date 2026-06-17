package withdrawalservice

import (
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

// ErrNotFound is used when a record is not found
var ErrNotFound = errors.New("record not found")

// mockWithdrawalRepository implements WithdrawalRepository for testing
type mockWithdrawalRepository struct {
	withdrawals map[int64]*repository.Withdrawal
	nextID      int64
}

func newMockWithdrawalRepository() *mockWithdrawalRepository {
	return &mockWithdrawalRepository{
		withdrawals: make(map[int64]*repository.Withdrawal),
		nextID:      1,
	}
}

func (m *mockWithdrawalRepository) Create(withdrawal *repository.Withdrawal) error {
	withdrawal.ID = m.nextID
	m.nextID++
	withdrawal.CreatedAt = time.Now()
	withdrawal.UpdatedAt = time.Now()
	m.withdrawals[withdrawal.ID] = withdrawal
	return nil
}

func (m *mockWithdrawalRepository) GetByID(id int64) (*repository.Withdrawal, error) {
	if w, ok := m.withdrawals[id]; ok {
		return w, nil
	}
	return nil, ErrNotFound
}

func (m *mockWithdrawalRepository) GetByIdempotencyKey(key string) (*repository.Withdrawal, error) {
	for _, w := range m.withdrawals {
		if w.IdempotencyKey == key {
			return w, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockWithdrawalRepository) GetByTxHash(chainID int64, txHash string) (*repository.Withdrawal, error) {
	for _, w := range m.withdrawals {
		if w.ChainID == chainID && w.TxHash == txHash {
			return w, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockWithdrawalRepository) Update(withdrawal *repository.Withdrawal) error {
	withdrawal.UpdatedAt = time.Now()
	m.withdrawals[withdrawal.ID] = withdrawal
	return nil
}

func (m *mockWithdrawalRepository) UpdateStatus(id int64, status repository.WithdrawalStatus) error {
	if w, ok := m.withdrawals[id]; ok {
		w.Status = status
		w.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (m *mockWithdrawalRepository) Delete(id int64) error {
	delete(m.withdrawals, id)
	return nil
}

func (m *mockWithdrawalRepository) List(limit int) ([]*repository.Withdrawal, error) {
	var result []*repository.Withdrawal
	for _, w := range m.withdrawals {
		result = append(result, w)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockWithdrawalRepository) ListByChainID(chainID int64, limit int) ([]*repository.Withdrawal, error) {
	var result []*repository.Withdrawal
	for _, w := range m.withdrawals {
		if w.ChainID == chainID {
			result = append(result, w)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockWithdrawalRepository) ListByFromAddress(chainID int64, fromAddress string, limit int) ([]*repository.Withdrawal, error) {
	var result []*repository.Withdrawal
	for _, w := range m.withdrawals {
		if w.ChainID == chainID && w.FromAddress == fromAddress {
			result = append(result, w)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockWithdrawalRepository) ListByStatus(status repository.WithdrawalStatus, limit int) ([]*repository.Withdrawal, error) {
	var result []*repository.Withdrawal
	for _, w := range m.withdrawals {
		if w.Status == status {
			result = append(result, w)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// mockBalanceRepository implements BalanceRepository for testing
type mockBalanceRepository struct{}

func (m *mockBalanceRepository) Create(balance *repository.Balance) error {
	return nil
}

func (m *mockBalanceRepository) GetByID(id int64) (*repository.Balance, error) {
	return nil, ErrNotFound
}

func (m *mockBalanceRepository) GetByAccountChainAndToken(account string, chainID, tokenID int64) (*repository.Balance, error) {
	return &repository.Balance{
		ID:               1,
		AccountAddress:   account,
		ChainID:          chainID,
		TokenID:          tokenID,
		AvailableBalance: "1000000000000000000",
		FrozenBalance:    "0",
	}, nil
}

func (m *mockBalanceRepository) Update(balance *repository.Balance) error {
	return nil
}

func (m *mockBalanceRepository) Credit(account string, chainID, tokenID int64, amount string) (*repository.Balance, error) {
	return &repository.Balance{
		AccountAddress:   account,
		ChainID:          chainID,
		TokenID:          tokenID,
		AvailableBalance: amount,
		FrozenBalance:    "0",
	}, nil
}

func (m *mockBalanceRepository) Freeze(account string, chainID, tokenID int64, amount string) error {
	return nil
}

func (m *mockBalanceRepository) Unfreeze(account string, chainID, tokenID int64, amount string) error {
	return nil
}

func (m *mockBalanceRepository) Delete(id int64) error {
	return nil
}

func (m *mockBalanceRepository) List(limit int) ([]*repository.Balance, error) {
	return nil, nil
}

func (m *mockBalanceRepository) ListByAccountAddress(account string) ([]*repository.Balance, error) {
	return nil, nil
}

func (m *mockBalanceRepository) ListByChainID(chainID int64) ([]*repository.Balance, error) {
	return nil, nil
}

// mockNatsClient implements a minimal NATS client for testing
type mockNatsClient struct {
	published []string
}

func (m *mockNatsClient) Publish(subject string, data []byte) error {
	m.published = append(m.published, subject)
	return nil
}

func (m *mockNatsClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return nil, nil
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		addr  string
		valid bool
	}{
		{"0x1234567890123456789012345678901234567890", true},
		{"0xabcDEF1234567890abcdef1234567890abcdef12", true},
		{"0x", false},
		{"0x123", false},
		{"1234567890123456789012345678901234567890", false},
		{"0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			result := isValidAddress(tt.addr)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestWithdrawalWorker_RiskCheck(t *testing.T) {
	mockRepo := newMockWithdrawalRepository()
	mockBalance := &mockBalanceRepository{}

	withdrawal := &repository.Withdrawal{
		ID:             1,
		ChainID:        56,
		TokenID:        1,
		FromAddress:    "0x1234567890123456789012345678901234567890",
		ToAddress:      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:         "1000000000000000000",
		Status:         repository.WithdrawalStatusCreated,
		IdempotencyKey: "test-key-1",
	}

	// Test valid withdrawal
	err := performRiskCheck(withdrawal)
	assert.NoError(t, err)

	// Test invalid address
	withdrawal.ToAddress = "invalid"
	err = performRiskCheck(withdrawal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid destination address")

	// Reset and test zero amount
	withdrawal.ToAddress = "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	withdrawal.Amount = "0"
	err = performRiskCheck(withdrawal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid amount")

	_ = mockRepo
	_ = mockBalance
}

func TestEvaluateRisk_AutoApprove(t *testing.T) {
	worker := &WithdrawalWorker{
		balanceRepo: &mockBalanceRepository{},
		logger:      slog.Default(),
		riskConfig: RiskConfig{
			HotWalletAddress:      normalizeAddress("0x1234567890123456789012345678901234567890"),
			MaxAutoApproveAmount:  big.NewInt(1000),
			RequireWhitelist:      true,
			AllowedDestinationSet: map[string]struct{}{normalizeAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): {}},
		},
	}

	decision := worker.evaluateRisk(&repository.Withdrawal{
		FromAddress: "0x1234567890123456789012345678901234567890",
		ToAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:      "100",
		ChainID:     56,
		TokenID:     1,
	})

	assert.Equal(t, repository.WithdrawalStatusApproved, decision.Status)
	assert.Empty(t, decision.Reason)
}

func TestEvaluateRisk_ManualReviewForWhitelist(t *testing.T) {
	worker := &WithdrawalWorker{
		balanceRepo: &mockBalanceRepository{},
		logger:      slog.Default(),
		riskConfig: RiskConfig{
			HotWalletAddress:      normalizeAddress("0x1234567890123456789012345678901234567890"),
			MaxAutoApproveAmount:  big.NewInt(1000),
			RequireWhitelist:      true,
			AllowedDestinationSet: map[string]struct{}{},
		},
	}

	decision := worker.evaluateRisk(&repository.Withdrawal{
		FromAddress: "0x1234567890123456789012345678901234567890",
		ToAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:      "100",
		ChainID:     56,
		TokenID:     1,
	})

	assert.Equal(t, repository.WithdrawalStatusManualReview, decision.Status)
	assert.Contains(t, decision.Reason, "whitelist")
}

func TestEvaluateRisk_ManualReviewForHighAmount(t *testing.T) {
	worker := &WithdrawalWorker{
		balanceRepo: &mockBalanceRepository{},
		logger:      slog.Default(),
		riskConfig: RiskConfig{
			HotWalletAddress:      normalizeAddress("0x1234567890123456789012345678901234567890"),
			MaxAutoApproveAmount:  big.NewInt(50),
			RequireWhitelist:      false,
			AllowedDestinationSet: map[string]struct{}{},
		},
	}

	decision := worker.evaluateRisk(&repository.Withdrawal{
		FromAddress: "0x1234567890123456789012345678901234567890",
		ToAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:      "100",
		ChainID:     56,
		TokenID:     1,
	})

	assert.Equal(t, repository.WithdrawalStatusManualReview, decision.Status)
	assert.Contains(t, decision.Reason, "auto-approval threshold")
}

func TestEvaluateRisk_FailedForInsufficientBalance(t *testing.T) {
	worker := &WithdrawalWorker{
		balanceRepo: &lowBalanceRepository{},
		logger:      slog.Default(),
		riskConfig: RiskConfig{
			HotWalletAddress:      normalizeAddress("0x1234567890123456789012345678901234567890"),
			MaxAutoApproveAmount:  big.NewInt(1000),
			RequireWhitelist:      false,
			AllowedDestinationSet: map[string]struct{}{},
		},
	}

	decision := worker.evaluateRisk(&repository.Withdrawal{
		FromAddress: "0x1234567890123456789012345678901234567890",
		ToAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:      "100",
		ChainID:     56,
		TokenID:     1,
	})

	assert.Equal(t, repository.WithdrawalStatusFailed, decision.Status)
	assert.Contains(t, decision.Reason, "insufficient")
}

func TestLoadRiskConfig(t *testing.T) {
	t.Setenv("WITHDRAWAL_HOT_WALLET_ADDRESS", "0x1234567890123456789012345678901234567890")
	t.Setenv("WITHDRAWAL_MAX_AUTO_APPROVE_AMOUNT", "999")
	t.Setenv("WITHDRAWAL_REQUIRE_WHITELIST", "true")
	t.Setenv("WITHDRAWAL_ALLOWED_DESTINATIONS", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111")

	cfg := LoadRiskConfig()
	assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.HotWalletAddress)
	assert.Equal(t, "999", cfg.MaxAutoApproveAmount.String())
	assert.True(t, cfg.RequireWhitelist)
	assert.Len(t, cfg.AllowedDestinationSet, 2)
}

// performRiskCheck is a helper that duplicates the risk check logic for testing
func performRiskCheck(withdrawal *repository.Withdrawal) error {
	if !isValidAddress(withdrawal.ToAddress) {
		return &riskCheckError{fmt: "invalid destination address: %s", args: []interface{}{withdrawal.ToAddress}}
	}
	if withdrawal.Amount == "" || withdrawal.Amount == "0" {
		return &riskCheckError{fmt: "invalid amount: %s", args: []interface{}{withdrawal.Amount}}
	}
	return nil
}

type riskCheckError struct {
	fmt  string
	args []interface{}
}

func (e *riskCheckError) Error() string {
	if len(e.args) == 0 {
		return e.fmt
	}
	return fmt.Sprintf(e.fmt, e.args...)
}

func TestWithdrawalBroadcastEvent(t *testing.T) {
	event := WithdrawalBroadcastEvent{
		WithdrawalID: 1,
		ChainID:      56,
		TokenID:      1,
		FromAddress:  "0x1234567890123456789012345678901234567890",
		ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:       "1000000000000000000",
		Nonce:        123,
	}

	assert.Equal(t, int64(1), event.WithdrawalID)
	assert.Equal(t, int64(56), event.ChainID)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", event.FromAddress)
	assert.Equal(t, "1000000000000000000", event.Amount)
	assert.Equal(t, int64(123), event.Nonce)
}

func TestWithdrawalApprovedEvent(t *testing.T) {
	event := WithdrawalApprovedEvent{
		WithdrawalID: 1,
		ChainID:      56,
		TokenID:      1,
		FromAddress:  "0x1234567890123456789012345678901234567890",
		ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Amount:       "1000000000000000000",
	}

	assert.Equal(t, int64(1), event.WithdrawalID)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", event.FromAddress)
}

// Ensure mock types implement interfaces
var _ repository.WithdrawalRepository = (*mockWithdrawalRepository)(nil)
var _ repository.BalanceRepository = (*mockBalanceRepository)(nil)

type lowBalanceRepository struct{}

func (m *lowBalanceRepository) Create(balance *repository.Balance) error { return nil }
func (m *lowBalanceRepository) GetByID(id int64) (*repository.Balance, error) {
	return nil, ErrNotFound
}
func (m *lowBalanceRepository) GetByAccountChainAndToken(account string, chainID, tokenID int64) (*repository.Balance, error) {
	return &repository.Balance{
		ID:               1,
		AccountAddress:   account,
		ChainID:          chainID,
		TokenID:          tokenID,
		AvailableBalance: "10",
		FrozenBalance:    "0",
	}, nil
}
func (m *lowBalanceRepository) Update(balance *repository.Balance) error { return nil }
func (m *lowBalanceRepository) Credit(account string, chainID, tokenID int64, amount string) (*repository.Balance, error) {
	return nil, nil
}
func (m *lowBalanceRepository) Freeze(account string, chainID, tokenID int64, amount string) error {
	return nil
}
func (m *lowBalanceRepository) Unfreeze(account string, chainID, tokenID int64, amount string) error {
	return nil
}
func (m *lowBalanceRepository) Delete(id int64) error                         { return nil }
func (m *lowBalanceRepository) List(limit int) ([]*repository.Balance, error) { return nil, nil }
func (m *lowBalanceRepository) ListByAccountAddress(account string) ([]*repository.Balance, error) {
	return nil, nil
}
func (m *lowBalanceRepository) ListByChainID(chainID int64) ([]*repository.Balance, error) {
	return nil, nil
}

var _ repository.BalanceRepository = (*lowBalanceRepository)(nil)
