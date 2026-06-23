package broadcaster

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EVMRPCAdapter struct {
	chainRepo    repository.ChainRepository
	providerRepo repository.RPCProviderRepository
	timeout      time.Duration
	retries      int
}

func NewEVMRPCAdapter(
	chainRepo repository.ChainRepository,
	providerRepo repository.RPCProviderRepository,
) *EVMRPCAdapter {
	return &EVMRPCAdapter{
		chainRepo:    chainRepo,
		providerRepo: providerRepo,
		timeout:      12 * time.Second,
		retries:      2,
	}
}

func (a *EVMRPCAdapter) GasPrice(ctx context.Context, chainID int64) (*big.Int, error) {
	var gasPrice *big.Int
	err := a.withClient(ctx, chainID, func(callCtx context.Context, client *ethclient.Client) error {
		var err error
		gasPrice, err = client.SuggestGasPrice(callCtx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return gasPrice, nil
}

func (a *EVMRPCAdapter) SendRawTransaction(ctx context.Context, chainID int64, signedTx []byte) (string, error) {
	var tx types.Transaction
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return "", fmt.Errorf("invalid signed transaction bytes: %w", err)
	}

	err := a.withClient(ctx, chainID, func(callCtx context.Context, client *ethclient.Client) error {
		return client.SendTransaction(callCtx, &tx)
	})
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (a *EVMRPCAdapter) NonceAt(ctx context.Context, chainID int64, address string) (uint64, error) {
	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("invalid address: %s", address)
	}

	var nonce uint64
	err := a.withClient(ctx, chainID, func(callCtx context.Context, client *ethclient.Client) error {
		var err error
		nonce, err = client.PendingNonceAt(callCtx, common.HexToAddress(address))
		return err
	})
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (a *EVMRPCAdapter) EstimateGas(ctx context.Context, req *EstimateGasRequest) (uint64, error) {
	if req == nil {
		return 0, fmt.Errorf("estimate gas request is required")
	}

	var gas uint64
	err := a.withClient(ctx, req.ChainID, func(callCtx context.Context, client *ethclient.Client) error {
		var err error
		gas, err = client.EstimateGas(callCtx, ethereum.CallMsg{
			From:     req.From,
			To:       req.To,
			GasPrice: req.GasPrice,
			Value:    req.Value,
			Data:     req.Data,
		})
		return err
	})
	if err != nil {
		return 0, err
	}

	return gas, nil
}

func (a *EVMRPCAdapter) GetTransactionReceipt(ctx context.Context, chainID int64, txHash string) (*TxReceipt, error) {
	if strings.TrimSpace(txHash) == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}

	var receipt *types.Receipt
	err := a.withClient(ctx, chainID, func(callCtx context.Context, client *ethclient.Client) error {
		var err error
		receipt, err = client.TransactionReceipt(callCtx, common.HexToHash(txHash))
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
			return rpcgateway.ErrReceiptNotFound
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return &TxReceipt{
		Status:      receipt.Status,
		BlockNumber: receipt.BlockNumber.Uint64(),
	}, nil
}

type EstimateGasRequest struct {
	ChainID  int64
	From     common.Address
	To       *common.Address
	Value    *big.Int
	GasPrice *big.Int
	Data     []byte
}

func (a *EVMRPCAdapter) withClient(
	ctx context.Context,
	chainID int64,
	fn func(context.Context, *ethclient.Client) error,
) error {
	if fn == nil {
		return fmt.Errorf("rpc callback is required")
	}

	chain, err := a.chainRepo.GetByID(chainID)
	if err == nil && chain != nil {
		chainID = chain.ID
	}

	opts := rpcgateway.DefaultCallOptions()
	opts.Timeout = a.timeout
	opts.MaxRetries = a.retries

	client, err := rpcgateway.NewResilientClient(chainID, a.providerRepo, opts)
	if err != nil {
		return fmt.Errorf("failed to build resilient rpc client: %w", err)
	}

	return client.Call(ctx, fn)
}
