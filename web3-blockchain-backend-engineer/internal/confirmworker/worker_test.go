package confirmworker

import (
	"encoding/json"
	"testing"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestMakeIdempotencyKey(t *testing.T) {
	key := MakeIdempotencyKey(56, "0xtxhash123", 0)
	assert.Equal(t, "56:0xtxhash123:0", key)
}

func TestParsedEventMessage_JSONParsing(t *testing.T) {
	// Verify ParsedEventMessage can parse JSON from parser correctly
	jsonData := `{
		"chain_id": 56,
		"token_id": 1,
		"token_address": "0xContract",
		"tx_hash": "0xtxhash",
		"log_index": 0,
		"block_number": 12345678,
		"block_hash": "0xblockhash",
		"from": "0xfrom",
		"to": "0xto",
		"amount": "1000000",
		"event_name": "Transfer"
	}`

	var msg ParsedEventMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	assert.NoError(t, err)

	assert.Equal(t, int64(56), msg.ChainID)
	assert.Equal(t, int64(1), msg.TokenID)
	assert.Equal(t, "0xContract", msg.TokenAddress)
	assert.Equal(t, "0xtxhash", msg.TxHash)
	assert.Equal(t, uint(0), msg.LogIndex)
	assert.Equal(t, int64(12345678), msg.BlockNumber)
	assert.Equal(t, "0xblockhash", msg.BlockHash)
	assert.Equal(t, "0xfrom", msg.From)
	assert.Equal(t, "0xto", msg.To)
	assert.Equal(t, "1000000", msg.Amount)
	assert.Equal(t, "Transfer", msg.EventName)
}

func TestDepositConfirmedEvent_JSONMarshaling(t *testing.T) {
	event := DepositConfirmedEvent{
		DepositID:   1,
		ChainID:     56,
		TokenID:     1,
		Account:     "0xto",
		Amount:      "1000000",
		TxHash:      "0xtxhash",
		BlockNumber: 12345678,
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), parsed["deposit_id"])
	assert.Equal(t, float64(56), parsed["chain_id"])
	assert.Equal(t, float64(1), parsed["token_id"])
	assert.Equal(t, "0xto", parsed["account"])
	assert.Equal(t, "1000000", parsed["amount"])
	assert.Equal(t, "0xtxhash", parsed["tx_hash"])
	assert.Equal(t, float64(12345678), parsed["block_number"])
}

func TestTrackedDeposit_PublishFailed(t *testing.T) {
	// Test that trackedDeposit can track publish failures for retry
	deposit := &repository.Deposit{
		ID:            1,
		ChainID:       56,
		TokenID:       1,
		TxHash:        "0xtxhash",
		BlockNumber:   12345678,
		Status:        repository.DepositStatusConfirmed,
		Confirmations: 12,
	}

	td := &trackedDeposit{
		Deposit:       deposit,
		RequiredConfs: 12,
		PublishFailed: false,
	}

	assert.False(t, td.PublishFailed)

	// Simulate publish failure
	td.PublishFailed = true
	assert.True(t, td.PublishFailed)
	assert.Equal(t, repository.DepositStatusConfirmed, td.Deposit.Status)
}

// Mock implementations

type mockDepositRepo struct {
	deposits map[int64]*repository.Deposit
}

func newMockDepositRepo() *mockDepositRepo {
	return &mockDepositRepo{
		deposits: make(map[int64]*repository.Deposit),
	}
}

func (m *mockDepositRepo) Create(deposit *repository.Deposit) error {
	deposit.ID = int64(len(m.deposits) + 1)
	m.deposits[deposit.ID] = deposit
	return nil
}

func (m *mockDepositRepo) GetByID(id int64) (*repository.Deposit, error) {
	if d, ok := m.deposits[id]; ok {
		return d, nil
	}
	return nil, nil
}

func (m *mockDepositRepo) GetByIdempotencyKey(key string) (*repository.Deposit, error) {
	for _, d := range m.deposits {
		if d.IdempotencyKey == key {
			return d, nil
		}
	}
	return nil, nil
}

func (m *mockDepositRepo) UpdateStatus(id int64, status repository.DepositStatus) error {
	if d, ok := m.deposits[id]; ok {
		d.Status = status
	}
	return nil
}

func (m *mockDepositRepo) SetConfirmations(id int64, confirmations int) error {
	if d, ok := m.deposits[id]; ok {
		d.Confirmations = confirmations
	}
	return nil
}

func (m *mockDepositRepo) ConfirmWithCondition(id int64, confirmations, targetConfirmations int) (bool, error) {
	if d, ok := m.deposits[id]; ok {
		if d.Status == repository.DepositStatusPendingConfirmation && confirmations >= targetConfirmations {
			d.Status = repository.DepositStatusConfirmed
			d.Confirmations = confirmations
			d.ConfirmedEventPublished = false
			return true, nil
		}
	}
	return false, nil
}

func (m *mockDepositRepo) MarkConfirmedEventPublished(id int64) error {
	if d, ok := m.deposits[id]; ok {
		d.ConfirmedEventPublished = true
	}
	return nil
}

func (m *mockDepositRepo) ListByStatus(status repository.DepositStatus, limit int) ([]*repository.Deposit, error) {
	var result []*repository.Deposit
	for _, d := range m.deposits {
		if d.Status == status {
			result = append(result, d)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockDepositRepo) ListConfirmedUnpublished(limit int) ([]*repository.Deposit, error) {
	var result []*repository.Deposit
	for _, d := range m.deposits {
		if d.Status == repository.DepositStatusConfirmed && !d.ConfirmedEventPublished {
			result = append(result, d)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

type mockChainRepo struct {
	chains map[int64]*repository.Chain
}

func newMockChainRepo() *mockChainRepo {
	return &mockChainRepo{
		chains: map[int64]*repository.Chain{
			56: {ID: 1, ChainID: 56, Name: "BSC", FinalityConfirmations: 12},
		},
	}
}

func (m *mockChainRepo) GetByChainID(chainID int64) (*repository.Chain, error) {
	if c, ok := m.chains[chainID]; ok {
		return c, nil
	}
	return nil, nil
}

type mockRPCClient struct {
	blockNumber uint64
}

func newMockRPCClient(blockNumber uint64) *mockRPCClient {
	return &mockRPCClient{blockNumber: blockNumber}
}

func (m *mockRPCClient) BlockNumber() (uint64, error) {
	return m.blockNumber, nil
}
