package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestDepositRepository_Create(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	deposit := &Deposit{
		ChainID:             1,
		TokenID:             1,
		TxHash:              "0xtxhash123",
		LogIndex:            0,
		FromAddress:         "0xfromaddress",
		ToAddress:           "0xtoaddress",
		Amount:              "1000000000000000000",
		BlockNumber:         12345678,
		Status:              DepositStatusDetected,
		Confirmations:       0,
		TargetConfirmations: 12,
		IdempotencyKey:      "1:0xtxhash123:0",
	}

	mock.ExpectQuery(`INSERT INTO deposits`).
		WithArgs(
			deposit.ChainID,
			deposit.TokenID,
			deposit.TxHash,
			deposit.LogIndex,
			deposit.FromAddress,
			deposit.ToAddress,
			deposit.Amount,
			deposit.BlockNumber,
			deposit.Status,
			deposit.Confirmations,
			deposit.TargetConfirmations,
			deposit.IdempotencyKey,
			deposit.ProcessedAt,
			deposit.ConfirmedEventPublished,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create(deposit)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deposit.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_GetByIdempotencyKey(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "chain_id", "token_id", "tx_hash", "log_index",
		"from_address", "to_address", "amount", "block_number",
		"status", "confirmations", "target_confirmations", "idempotency_key",
		"processed_at", "confirmed_event_published", "created_at", "updated_at",
	}).AddRow(
		1, 1, 1, "0xtxhash123", 0,
		"0xfrom", "0xto", "1000000", 12345678,
		"detected", 0, 12, "1:0xtxhash123:0",
		nil, false, now, now,
	)

	mock.ExpectQuery(`SELECT .+ FROM deposits WHERE idempotency_key`).
		WithArgs("1:0xtxhash123:0").
		WillReturnRows(rows)

	deposit, err := repo.GetByIdempotencyKey("1:0xtxhash123:0")
	assert.NoError(t, err)
	assert.NotNil(t, deposit)
	assert.Equal(t, int64(1), deposit.ID)
	assert.Equal(t, "0xtxhash123", deposit.TxHash)
	assert.Equal(t, DepositStatusDetected, deposit.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_GetByIdempotencyKey_NotFound(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM deposits WHERE idempotency_key`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	deposit, err := repo.GetByIdempotencyKey("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, deposit)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_UpdateStatus(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	mock.ExpectExec(`UPDATE deposits SET status`).
		WithArgs(int64(1), DepositStatusConfirmed, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(1, DepositStatusConfirmed)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_SetConfirmations(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	mock.ExpectExec(`UPDATE deposits SET confirmations`).
		WithArgs(int64(1), 12, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetConfirmations(1, 12)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_ConfirmWithCondition_Success(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	mock.ExpectExec(`UPDATE deposits`).
		WithArgs(
			int64(1),
			DepositStatusConfirmed,
			12,
			sqlmock.AnyArg(),
			DepositStatusPendingConfirmation,
			12,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updated, err := repo.ConfirmWithCondition(1, 12, 12)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_ConfirmWithCondition_NotMet(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	// Status is still 'detected', not 'pending_confirmation'
	mock.ExpectExec(`UPDATE deposits SET status`).
		WithArgs(
			int64(1),
			DepositStatusConfirmed,
			12,
			sqlmock.AnyArg(),
			DepositStatusPendingConfirmation,
			12,
		).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	updated, err := repo.ConfirmWithCondition(1, 12, 12)
	assert.NoError(t, err)
	assert.False(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDepositRepository_ListByStatus(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewDepositRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "chain_id", "token_id", "tx_hash", "log_index",
		"from_address", "to_address", "amount", "block_number",
		"status", "confirmations", "target_confirmations", "idempotency_key",
		"processed_at", "confirmed_event_published", "created_at", "updated_at",
	}).
		AddRow(1, 1, 1, "0xtx1", 0, "0xfrom1", "0xto1", "100", 123, "detected", 0, 12, "key1", nil, false, now, now).
		AddRow(2, 1, 1, "0xtx2", 0, "0xfrom2", "0xto2", "200", 124, "detected", 0, 12, "key2", nil, false, now, now)

	mock.ExpectQuery(`SELECT .+ FROM deposits WHERE status`).
		WithArgs("detected", 100).
		WillReturnRows(rows)

	deposits, err := repo.ListByStatus(DepositStatusDetected, 100)
	assert.NoError(t, err)
	assert.Len(t, deposits, 2)
	assert.Equal(t, "0xtx1", deposits[0].TxHash)
	assert.Equal(t, "0xtx2", deposits[1].TxHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_Credit_NewBalance(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewBalanceRepository(db)

	// Balance doesn't exist, INSERT happens
	mock.ExpectQuery(`INSERT INTO balances .+ ON CONFLICT`).
		WithArgs("0xaccount", int64(1), int64(1), "1000000", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "available_balance", "frozen_balance"}).
			AddRow(1, "1000000", "0"))

	balance, err := repo.Credit("0xaccount", 1, 1, "1000000")
	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "0xaccount", balance.AccountAddress)
	assert.Equal(t, "1000000", balance.AvailableBalance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_Credit_ExistingBalance(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewBalanceRepository(db)

	// Balance exists, UPDATE happens
	mock.ExpectQuery(`INSERT INTO balances .+ ON CONFLICT`).
		WithArgs("0xaccount", int64(1), int64(1), "1000000", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "available_balance", "frozen_balance"}).
			AddRow(1, "2000000", "0")) // 1000000 + 1000000

	balance, err := repo.Credit("0xaccount", 1, 1, "1000000")
	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "2000000", balance.AvailableBalance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetByAccountChainAndToken(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewBalanceRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "account_address", "chain_id", "token_id", "available_balance", "frozen_balance", "updated_at",
	}).AddRow(1, "0xaccount", 1, 1, "1000000", "0", time.Now())

	mock.ExpectQuery(`SELECT .+ FROM balances WHERE account_address`).
		WithArgs("0xaccount", int64(1), int64(1)).
		WillReturnRows(rows)

	balance, err := repo.GetByAccountChainAndToken("0xaccount", 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "0xaccount", balance.AccountAddress)
	assert.Equal(t, "1000000", balance.AvailableBalance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetByAccountChainAndToken_NotFound(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewBalanceRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM balances WHERE account_address`).
		WithArgs("0xaccount", int64(1), int64(1)).
		WillReturnError(sql.ErrNoRows)

	balance, err := repo.GetByAccountChainAndToken("0xaccount", 1, 1)
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLedgerEntryRepository_Create(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewLedgerEntryRepository(db)

	entry := &LedgerEntry{
		AccountAddress: "0xaccount",
		ChainID:        1,
		TokenID:        1,
		Direction:      LedgerDirectionCredit,
		Amount:         "1000000",
		BalanceBefore:  "0",
		BalanceAfter:   "1000000",
		EntryType:      LedgerEntryTypeDeposit,
		ReferenceType:  ReferenceTypeDeposit,
		ReferenceID:    1,
	}

	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			entry.AccountAddress,
			entry.ChainID,
			entry.TokenID,
			entry.Direction,
			entry.Amount,
			entry.BalanceBefore,
			entry.BalanceAfter,
			entry.EntryType,
			entry.ReferenceType,
			entry.ReferenceID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create(entry)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), entry.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLedgerEntryRepository_ListByReference(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewLedgerEntryRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "account_address", "chain_id", "token_id", "direction",
		"amount", "balance_before", "balance_after", "entry_type",
		"reference_type", "reference_id", "reversal_of", "created_at",
	}).AddRow(1, "0xaccount", 1, 1, "credit", "1000000", "0", "1000000", "deposit", "deposit", 1, nil, now)

	mock.ExpectQuery(`SELECT .+ FROM ledger_entries WHERE reference_type`).
		WithArgs("deposit", int64(1)).
		WillReturnRows(rows)

	entries, err := repo.ListByReference(ReferenceTypeDeposit, 1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "1000000", entries[0].Amount)
	assert.Equal(t, "0", entries[0].BalanceBefore)
	assert.Equal(t, "1000000", entries[0].BalanceAfter)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainRepository_GetByChainID(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewChainRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "chain_id", "name", "native_symbol", "finality_confirmations", "is_active", "created_at",
	}).AddRow(1, 56, "BSC", "BNB", 12, true, time.Now())

	mock.ExpectQuery(`SELECT .+ FROM chains WHERE chain_id`).
		WithArgs(int64(56)).
		WillReturnRows(rows)

	chain, err := repo.GetByChainID(56)
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, "BSC", chain.Name)
	assert.Equal(t, "BNB", chain.NativeSymbol)
	assert.Equal(t, 12, chain.FinalityConfirmations)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetByChainIDAndContract(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewTokenRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "chain_id", "contract_address", "symbol", "decimals", "is_native", "is_active",
	}).AddRow(1, 56, "0xContract", "USDT", 18, false, true)

	mock.ExpectQuery(`SELECT .+ FROM tokens WHERE chain_id`).
		WithArgs(int64(56), "0xContract").
		WillReturnRows(rows)

	token, err := repo.GetByChainIDAndContract(56, "0xContract")
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "USDT", token.Symbol)
	assert.Equal(t, 18, token.Decimals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlockRepository_Create(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewBlockRepository(db)

	block := &Block{
		ChainID:     56,
		BlockNumber: 12345678,
		BlockHash:   "0xblockhash",
		ParentHash:  "0xparenthash",
		IsOrphaned:  false,
	}

	mock.ExpectQuery(`INSERT INTO blocks`).
		WithArgs(
			block.ChainID,
			block.BlockNumber,
			block.BlockHash,
			block.ParentHash,
			block.BlockTime,
			block.IsOrphaned,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create(block)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), block.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScanCheckpointRepository_UpdateLastScannedBlock(t *testing.T) {
	db, mock := NewMock(t)
	defer db.Close()
	repo := NewScanCheckpointRepository(db)

	mock.ExpectExec(`UPDATE scan_checkpoints SET last_scanned_block`).
		WithArgs(int64(56), int64(12345600), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateLastScannedBlock(56, 12345600)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
