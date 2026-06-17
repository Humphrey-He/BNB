package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestWithdrawalCreationService_Create_Success(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalCreationService(db)
	withdrawal := &repository.Withdrawal{
		ChainID:        1,
		TokenID:        2,
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		Amount:         "100",
		Status:         repository.WithdrawalStatusCreated,
		IdempotencyKey: "idem-1",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnRows(sqlmock.NewRows([]string{"balance_before", "balance_after"}).AddRow("1000", "900"))
	mock.ExpectQuery(`INSERT INTO withdrawals`).
		WithArgs(
			withdrawal.ChainID,
			withdrawal.TokenID,
			withdrawal.FromAddress,
			withdrawal.ToAddress,
			withdrawal.Amount,
			withdrawal.Status,
			sql.NullString{},
			sql.NullInt64{},
			withdrawal.IdempotencyKey,
			sql.NullString{},
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			withdrawal.FromAddress,
			withdrawal.ChainID,
			withdrawal.TokenID,
			repository.LedgerDirectionDebit,
			withdrawal.Amount,
			"1000",
			"900",
			repository.LedgerEntryTypeFreeze,
			repository.ReferenceTypeWithdrawal,
			int64(10),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectCommit()

	err := service.Create(context.Background(), withdrawal)
	require.NoError(t, err)
	assert.Equal(t, int64(10), withdrawal.ID)
	assert.False(t, withdrawal.CreatedAt.IsZero())
	assert.False(t, withdrawal.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawalCreationService_Create_InsufficientBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalCreationService(db)
	withdrawal := &repository.Withdrawal{
		ChainID:        1,
		TokenID:        2,
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		Amount:         "100",
		Status:         repository.WithdrawalStatusCreated,
		IdempotencyKey: "idem-2",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := service.Create(context.Background(), withdrawal)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient available balance")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawalCreationService_Create_RollbackOnLedgerFailure(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalCreationService(db)
	withdrawal := &repository.Withdrawal{
		ChainID:        1,
		TokenID:        2,
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		Amount:         "100",
		Status:         repository.WithdrawalStatusCreated,
		IdempotencyKey: "idem-3",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnRows(sqlmock.NewRows([]string{"balance_before", "balance_after"}).AddRow("1000", "900"))
	mock.ExpectQuery(`INSERT INTO withdrawals`).
		WithArgs(
			withdrawal.ChainID,
			withdrawal.TokenID,
			withdrawal.FromAddress,
			withdrawal.ToAddress,
			withdrawal.Amount,
			withdrawal.Status,
			sql.NullString{},
			sql.NullInt64{},
			withdrawal.IdempotencyKey,
			sql.NullString{},
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			withdrawal.FromAddress,
			withdrawal.ChainID,
			withdrawal.TokenID,
			repository.LedgerDirectionDebit,
			withdrawal.Amount,
			"1000",
			"900",
			repository.LedgerEntryTypeFreeze,
			repository.ReferenceTypeWithdrawal,
			int64(11),
			sqlmock.AnyArg(),
		).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := service.Create(context.Background(), withdrawal)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "freeze ledger entry")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawalReviewService_Approve_Success(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalReviewService(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at`).
		WithArgs(int64(20)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "chain_id", "token_id", "from_address", "to_address", "amount", "status", "tx_hash", "nonce", "idempotency_key", "failure_reason", "created_at", "updated_at",
		}).AddRow(20, 1, 2, "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222", "100", repository.WithdrawalStatusManualReview, nil, nil, "idem-20", "need review", now, now))
	mock.ExpectExec(`UPDATE withdrawals`).
		WithArgs(int64(20), repository.WithdrawalStatusApproved, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	withdrawal, err := service.Approve(context.Background(), 20)
	require.NoError(t, err)
	assert.Equal(t, repository.WithdrawalStatusApproved, withdrawal.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawalReviewService_Reject_Success(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalReviewService(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at`).
		WithArgs(int64(21)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "chain_id", "token_id", "from_address", "to_address", "amount", "status", "tx_hash", "nonce", "idempotency_key", "failure_reason", "created_at", "updated_at",
		}).AddRow(21, 1, 2, "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222", "100", repository.WithdrawalStatusManualReview, nil, nil, "idem-21", "need review", now, now))
	mock.ExpectQuery(`SELECT id\s+FROM ledger_entries`).
		WithArgs(repository.ReferenceTypeWithdrawal, int64(21), repository.LedgerEntryTypeUnfreeze).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs("0x1111111111111111111111111111111111111111", int64(1), int64(2), "100").
		WillReturnRows(sqlmock.NewRows([]string{"available_before", "available_after"}).AddRow("900", "1000"))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			"0x1111111111111111111111111111111111111111",
			int64(1),
			int64(2),
			repository.LedgerDirectionCredit,
			"100",
			"900",
			"1000",
			repository.LedgerEntryTypeUnfreeze,
			repository.ReferenceTypeWithdrawal,
			int64(21),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`UPDATE withdrawals`).
		WithArgs(int64(21), repository.WithdrawalStatusCanceled, "risk rejected", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	withdrawal, err := service.Reject(context.Background(), 21, "risk rejected")
	require.NoError(t, err)
	assert.Equal(t, repository.WithdrawalStatusCanceled, withdrawal.Status)
	assert.Equal(t, "risk rejected", withdrawal.FailureReason)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawalReviewService_Approve_ConflictStatus(t *testing.T) {
	db, mock := newSQLMock(t)
	defer db.Close()

	service := newWithdrawalReviewService(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at`).
		WithArgs(int64(22)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "chain_id", "token_id", "from_address", "to_address", "amount", "status", "tx_hash", "nonce", "idempotency_key", "failure_reason", "created_at", "updated_at",
		}).AddRow(22, 1, 2, "0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222", "100", repository.WithdrawalStatusApproved, nil, nil, "idem-22", nil, now, now))
	mock.ExpectRollback()

	withdrawal, err := service.Approve(context.Background(), 22)
	require.Error(t, err)
	assert.Nil(t, withdrawal)
	assert.Contains(t, err.Error(), "manual_review")
	assert.NoError(t, mock.ExpectationsWereMet())
}
