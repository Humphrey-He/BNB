package broadcaster

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const erc20ABIJSON = `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`

type Signer interface {
	Address() common.Address
	SignWithdrawal(ctx context.Context, req *SignRequest) ([]byte, error)
}

type SignRequest struct {
	ChainID              int64
	Nonce                uint64
	To                   common.Address
	Value                *big.Int
	GasLimit             uint64
	GasPrice             *big.Int
	Token                *repository.Token
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
}

type LocalHexKeySigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	erc20ABI   abi.ABI
}

func NewLocalHexKeySigner(hexKey string) (*LocalHexKeySigner, error) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(hexKey, "0x"))
	if trimmed == "" {
		return nil, fmt.Errorf("empty signer private key")
	}

	privateKey, err := crypto.HexToECDSA(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid signer private key: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse erc20 abi: %w", err)
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to derive signer public key")
	}

	return &LocalHexKeySigner{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(*publicKey),
		erc20ABI:   parsedABI,
	}, nil
}

func (s *LocalHexKeySigner) Address() common.Address {
	return s.address
}

func (s *LocalHexKeySigner) SignWithdrawal(ctx context.Context, req *SignRequest) ([]byte, error) {
	_ = ctx
	if req == nil {
		return nil, fmt.Errorf("sign request is required")
	}
	if req.Value == nil {
		return nil, fmt.Errorf("withdrawal amount is required")
	}
	if req.GasPrice == nil && (req.MaxFeePerGas == nil || req.MaxPriorityFeePerGas == nil) {
		return nil, fmt.Errorf("gas price or eip1559 fees are required")
	}

	chainID := big.NewInt(req.ChainID)
	tx, err := s.buildTransaction(req)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	raw, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode signed transaction: %w", err)
	}

	return raw, nil
}

func (s *LocalHexKeySigner) buildTransaction(req *SignRequest) (*types.Transaction, error) {
	if req.Token == nil || req.Token.IsNative {
		return s.buildNativeTransfer(req), nil
	}

	if !common.IsHexAddress(req.Token.ContractAddress) {
		return nil, fmt.Errorf("invalid token contract address: %s", req.Token.ContractAddress)
	}

	input, err := s.erc20ABI.Pack("transfer", req.To, req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode erc20 transfer: %w", err)
	}

	if req.MaxFeePerGas != nil && req.MaxPriorityFeePerGas != nil {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(req.ChainID),
			Nonce:     req.Nonce,
			To:        ptrAddress(common.HexToAddress(req.Token.ContractAddress)),
			Value:     big.NewInt(0),
			Gas:       req.GasLimit,
			GasFeeCap: req.MaxFeePerGas,
			GasTipCap: req.MaxPriorityFeePerGas,
			Data:      input,
		}), nil
	}

	return types.NewTransaction(
		req.Nonce,
		common.HexToAddress(req.Token.ContractAddress),
		big.NewInt(0),
		req.GasLimit,
		req.GasPrice,
		input,
	), nil
}

func (s *LocalHexKeySigner) buildNativeTransfer(req *SignRequest) *types.Transaction {
	if req.MaxFeePerGas != nil && req.MaxPriorityFeePerGas != nil {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(req.ChainID),
			Nonce:     req.Nonce,
			To:        ptrAddress(req.To),
			Value:     req.Value,
			Gas:       req.GasLimit,
			GasFeeCap: req.MaxFeePerGas,
			GasTipCap: req.MaxPriorityFeePerGas,
		})
	}

	return types.NewTransaction(
		req.Nonce,
		req.To,
		req.Value,
		req.GasLimit,
		req.GasPrice,
		nil,
	)
}

func ptrAddress(addr common.Address) *common.Address {
	return &addr
}
