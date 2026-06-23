package broadcaster

import (
	"context"
	"math/big"
	"testing"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestLocalHexKeySigner_SignNativeWithdrawal(t *testing.T) {
	signer, err := NewLocalHexKeySigner("4c0883a69102937d6231471b5dbb6204fe512961708279b1a7d3b5edecf4d1f4")
	require.NoError(t, err)

	raw, err := signer.SignWithdrawal(context.Background(), &SignRequest{
		ChainID:  11155111,
		Nonce:    7,
		To:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Value:    big.NewInt(1_000_000_000_000_000),
		GasLimit: 21_000,
		GasPrice: big.NewInt(2_000_000_000),
		Token: &repository.Token{
			ID:       1,
			IsNative: true,
		},
	})
	require.NoError(t, err)

	var tx types.Transaction
	require.NoError(t, tx.UnmarshalBinary(raw))
	require.Equal(t, uint64(7), tx.Nonce())
	require.Equal(t, uint64(21_000), tx.Gas())
	require.Equal(t, "0x1111111111111111111111111111111111111111", tx.To().Hex())
	require.Equal(t, 0, tx.Value().Cmp(big.NewInt(1_000_000_000_000_000)))
}

func TestLocalHexKeySigner_SignERC20Withdrawal(t *testing.T) {
	signer, err := NewLocalHexKeySigner("4c0883a69102937d6231471b5dbb6204fe512961708279b1a7d3b5edecf4d1f4")
	require.NoError(t, err)

	raw, err := signer.SignWithdrawal(context.Background(), &SignRequest{
		ChainID:  11155111,
		Nonce:    9,
		To:       common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Value:    big.NewInt(25_000_000),
		GasLimit: 70_000,
		GasPrice: big.NewInt(3_000_000_000),
		Token: &repository.Token{
			ID:              2,
			ContractAddress: "0x3333333333333333333333333333333333333333",
			IsNative:        false,
		},
	})
	require.NoError(t, err)

	var tx types.Transaction
	require.NoError(t, tx.UnmarshalBinary(raw))
	require.Equal(t, uint64(9), tx.Nonce())
	require.Equal(t, uint64(70_000), tx.Gas())
	require.Equal(t, "0x3333333333333333333333333333333333333333", tx.To().Hex())
	require.NotEmpty(t, tx.Data())
	require.Zero(t, tx.Value().Sign())
}
