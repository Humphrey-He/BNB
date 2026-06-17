package parser

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestMakeIdempotencyKey(t *testing.T) {
	key := MakeIdempotencyKey(56, "0xtxhash123", 0)
	assert.Equal(t, "56:0xtxhash123:0", key)
}

func TestParsedEvent_JSONMarshaling(t *testing.T) {
	// This tests the fix for the JSON field mismatch issue
	// ParsedEvent must have snake_case JSON tags to match ConfirmWorker's expectations
	event := ParsedEvent{
		ChainID:      56,
		TokenID:      1,
		TokenAddress: "0xContract",
		TxHash:       "0xtxhash",
		LogIndex:     0,
		BlockNumber:  12345678,
		BlockHash:    "0xblockhash",
		From:         "0xfrom",
		To:           "0xto",
		Amount:       "1000000000000000000",
		EventName:    "Transfer",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	assert.NoError(t, err)

	// Unmarshal to a map to verify snake_case keys
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	// Verify all fields are correctly marshaled with snake_case keys
	assert.Equal(t, float64(56), parsed["chain_id"])
	assert.Equal(t, "0xContract", parsed["token_address"])
	assert.Equal(t, "0xtxhash", parsed["tx_hash"])
	assert.Equal(t, "Transfer", parsed["event_name"])
	assert.Equal(t, "0xto", parsed["to"])
	assert.Equal(t, "0xfrom", parsed["from"])
}

func TestParsedEventMessage_JSONParsing(t *testing.T) {
	// Test that we can parse JSON with snake_case into ParsedEvent
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

	// Parse JSON into ParsedEvent
	var event ParsedEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	assert.NoError(t, err)

	assert.Equal(t, int64(56), event.ChainID)
	assert.Equal(t, int64(1), event.TokenID)
	assert.Equal(t, "0xContract", event.TokenAddress)
	assert.Equal(t, "0xtxhash", event.TxHash)
	assert.Equal(t, uint(0), event.LogIndex)
	assert.Equal(t, int64(12345678), event.BlockNumber)
	assert.Equal(t, "0xblockhash", event.BlockHash)
	assert.Equal(t, "0xfrom", event.From)
	assert.Equal(t, "0xto", event.To)
	assert.Equal(t, "1000000", event.Amount)
	assert.Equal(t, "Transfer", event.EventName)
}

// Mock implementations for testing

type mockWatchedAddressRepo struct {
	watched map[string]*repository.WatchedAddress
}

func newMockWatchedAddressRepo() *mockWatchedAddressRepo {
	return &mockWatchedAddressRepo{
		watched: map[string]*repository.WatchedAddress{
			"56:0xto": {ID: 1, ChainID: 56, Address: "0xto", IsActive: true},
		},
	}
}

func (m *mockWatchedAddressRepo) GetByChainIDAndAddress(chainID int64, address string) (*repository.WatchedAddress, error) {
	key := "56:0xto"
	if addr, ok := m.watched[key]; ok {
		return addr, nil
	}
	return nil, errors.New("not found")
}

type mockTokenRepo struct {
	tokens map[string]*repository.Token
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		tokens: map[string]*repository.Token{
			"56:0xContract": {ID: 1, ChainID: 56, ContractAddress: "0xContract", Symbol: "USDT", IsActive: true},
		},
	}
}

func (m *mockTokenRepo) GetByChainIDAndContract(chainID int64, contractAddress string) (*repository.Token, error) {
	key := "56:0xContract"
	if token, ok := m.tokens[key]; ok {
		return token, nil
	}
	return nil, errors.New("not found")
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
	if chain, ok := m.chains[chainID]; ok {
		return chain, nil
	}
	return nil, errors.New("not found")
}

type mockDepositRepo struct {
	deposits map[string]*repository.Deposit
}

func newMockDepositRepo() *mockDepositRepo {
	return &mockDepositRepo{
		deposits: make(map[string]*repository.Deposit),
	}
}

func (m *mockDepositRepo) Create(deposit *repository.Deposit) error {
	m.deposits[deposit.IdempotencyKey] = deposit
	deposit.ID = int64(len(m.deposits))
	return nil
}

func (m *mockDepositRepo) GetByIdempotencyKey(key string) (*repository.Deposit, error) {
	if d, ok := m.deposits[key]; ok {
		return d, nil
	}
	return nil, errors.New("not found")
}

type mockChainEventRepo struct{}

func (m *mockChainEventRepo) Create(event *repository.ChainEvent) error {
	event.ID = 1
	return nil
}
