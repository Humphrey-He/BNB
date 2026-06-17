package broadcaster

import (
	"context"
	"fmt"
	"math/big"

	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
)

type ReceiptRPCAdapter struct {
	Client rpcgateway.Client
}

func (a *ReceiptRPCAdapter) GasPrice(ctx context.Context, chainID int64) (*big.Int, error) {
	return nil, fmt.Errorf("GasPrice is not implemented in ReceiptRPCAdapter")
}

func (a *ReceiptRPCAdapter) SendRawTransaction(ctx context.Context, chainID int64, signedTx []byte) (string, error) {
	return "", fmt.Errorf("SendRawTransaction is not implemented in ReceiptRPCAdapter")
}

func (a *ReceiptRPCAdapter) NonceAt(ctx context.Context, chainID int64, address string) (uint64, error) {
	return 0, fmt.Errorf("NonceAt is not implemented in ReceiptRPCAdapter")
}

func (a *ReceiptRPCAdapter) GetTransactionReceipt(ctx context.Context, chainID int64, txHash string) (*TxReceipt, error) {
	if a.Client == nil {
		return nil, fmt.Errorf("receipt rpc client is not configured")
	}

	receipt, err := a.Client.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	blockNumber := uint64(0)
	if receipt.BlockNumber != nil {
		blockNumber = receipt.BlockNumber.Uint64()
	}

	return &TxReceipt{
		Status:      receipt.Status,
		BlockNumber: blockNumber,
	}, nil
}
