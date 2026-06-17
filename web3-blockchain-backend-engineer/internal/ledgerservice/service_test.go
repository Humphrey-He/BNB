package ledgerservice

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepositConfirmedEvent_JSONParsing(t *testing.T) {
	jsonData := `{
		"deposit_id": 1,
		"chain_id": 56,
		"token_id": 1,
		"account": "0xto",
		"amount": "1000000",
		"tx_hash": "0xtxhash",
		"block_number": 12345678
	}`

	var event DepositConfirmedEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), event.DepositID)
	assert.Equal(t, int64(56), event.ChainID)
	assert.Equal(t, int64(1), event.TokenID)
	assert.Equal(t, "0xto", event.Account)
	assert.Equal(t, "1000000", event.Amount)
	assert.Equal(t, "0xtxhash", event.TxHash)
	assert.Equal(t, int64(12345678), event.BlockNumber)
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

	var parsed DepositConfirmedEvent
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, event.DepositID, parsed.DepositID)
	assert.Equal(t, event.ChainID, parsed.ChainID)
	assert.Equal(t, event.TokenID, parsed.TokenID)
	assert.Equal(t, event.Account, parsed.Account)
	assert.Equal(t, event.Amount, parsed.Amount)
	assert.Equal(t, event.TxHash, parsed.TxHash)
	assert.Equal(t, event.BlockNumber, parsed.BlockNumber)
}

func TestSubtractStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{"simple subtraction", "100", "50", "50"},
		{"zero result", "50", "50", "0"},
		{"negative becomes zero", "30", "50", "0"},
		{"large numbers", "1000000000", "1", "999999999"},
		{"zero a", "0", "0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtractStrings(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test helpers (would use testcontainers in real integration tests)
func TestLedgerService_Integration(t *testing.T) {
	// This is a placeholder for integration tests that would use testcontainers
	// to spin up a real PostgreSQL database
	//
	// Example structure:
	// func TestLedgerService_ProcessDepositCredit(t *testing.T) {
	//     ctx := context.Background()
	//     postgresContainer, err := testcontainerspostgres.Run(ctx, ...)
	//     defer func() { ... }()
	//
	//     db, err := sql.Open("postgres", connectionString)
	//     defer db.Close()
	//
	//     ledgerRepo := repository.NewLedgerEntryRepository(db)
	//     balanceRepo := repository.NewBalanceRepository(db)
	//     ...
	//
	//     service := NewLedgerService(db, natsClient, ledgerRepo, balanceRepo, depositRepo, logger)
	//     ...
	// }
	t.Skip("Integration test requires testcontainers - run manually with 'go test -v -tags=integration'")
}
