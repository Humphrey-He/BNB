package broadcaster

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestSettleBroadcastedWithdrawal_Success(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	withdrawal := &repository.Withdrawal{
		ID:          9,
		ChainID:     1,
		TokenID:     2,
		FromAddress: "0x1111111111111111111111111111111111111111",
		Amount:      "100",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id\s+FROM ledger_entries`).
		WithArgs(repository.ReferenceTypeWithdrawal, withdrawal.ID, repository.LedgerEntryTypeWithdrawal).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnRows(sqlmock.NewRows([]string{"frozen_before", "frozen_after"}).AddRow("100", "0"))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			withdrawal.FromAddress,
			withdrawal.ChainID,
			withdrawal.TokenID,
			repository.LedgerDirectionDebit,
			withdrawal.Amount,
			"100",
			"0",
			repository.LedgerEntryTypeWithdrawal,
			repository.ReferenceTypeWithdrawal,
			withdrawal.ID,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := settleBroadcastedWithdrawal(context.Background(), db, withdrawal)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseFailedWithdrawal_Success(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	withdrawal := &repository.Withdrawal{
		ID:          10,
		ChainID:     1,
		TokenID:     2,
		FromAddress: "0x1111111111111111111111111111111111111111",
		Amount:      "100",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id\s+FROM ledger_entries`).
		WithArgs(repository.ReferenceTypeWithdrawal, withdrawal.ID, repository.LedgerEntryTypeUnfreeze).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnRows(sqlmock.NewRows([]string{"available_before", "available_after"}).AddRow("900", "1000"))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			withdrawal.FromAddress,
			withdrawal.ChainID,
			withdrawal.TokenID,
			repository.LedgerDirectionCredit,
			withdrawal.Amount,
			"900",
			"1000",
			repository.LedgerEntryTypeUnfreeze,
			repository.ReferenceTypeWithdrawal,
			withdrawal.ID,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	err := releaseFailedWithdrawal(context.Background(), db, withdrawal)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSettlementIdempotentSkip(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	withdrawal := &repository.Withdrawal{ID: 11, ChainID: 1, TokenID: 2, FromAddress: "0x1111111111111111111111111111111111111111", Amount: "100"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id\s+FROM ledger_entries`).
		WithArgs(repository.ReferenceTypeWithdrawal, withdrawal.ID, repository.LedgerEntryTypeWithdrawal).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	err := settleBroadcastedWithdrawal(context.Background(), db, withdrawal)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompensateRevertedWithdrawal_Success(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	withdrawal := &repository.Withdrawal{
		ID:          12,
		ChainID:     1,
		TokenID:     2,
		FromAddress: "0x1111111111111111111111111111111111111111",
		Amount:      "100",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id\s+FROM ledger_entries`).
		WithArgs(repository.ReferenceTypeWithdrawal, withdrawal.ID, repository.LedgerEntryTypeReversal).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`UPDATE balances`).
		WithArgs(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).
		WillReturnRows(sqlmock.NewRows([]string{"balance_before", "balance_after"}).AddRow("900", "1000"))
	mock.ExpectQuery(`INSERT INTO ledger_entries`).
		WithArgs(
			withdrawal.FromAddress,
			withdrawal.ChainID,
			withdrawal.TokenID,
			repository.LedgerDirectionCredit,
			withdrawal.Amount,
			"900",
			"1000",
			repository.LedgerEntryTypeReversal,
			repository.ReferenceTypeWithdrawal,
			withdrawal.ID,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	err := compensateRevertedWithdrawal(context.Background(), db, withdrawal)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
