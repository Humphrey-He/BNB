package parser

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ERC-20 Transfer event signature
// keccak256("Transfer(address,address,uint256)")
var TransferEventSignature = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// DecodedTransfer represents a decoded ERC-20 Transfer event
type DecodedTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// DecodeTransfer decodes an ERC-20 Transfer event from a log
func DecodeTransfer(log *types.Log) (*DecodedTransfer, error) {
	// Verify this is a Transfer event
	if len(log.Topics) < 3 {
		return nil, errors.New("invalid log topics: expected at least 3 topics for Transfer event")
	}

	// Verify the event signature matches Transfer(address,address,uint256)
	if log.Topics[0] != TransferEventSignature {
		return nil, fmt.Errorf("event signature mismatch: expected %s, got %s", TransferEventSignature.Hex(), log.Topics[0].Hex())
	}

	// Extract from address (topics[1])
	from := common.HexToAddress(log.Topics[1].Hex())

	// Extract to address (topics[2])
	to := common.HexToAddress(log.Topics[2].Hex())

	// Extract amount from data
	if len(log.Data) == 0 {
		return nil, errors.New("invalid log data: expected token amount data")
	}

	amount := new(big.Int).SetBytes(log.Data)

	return &DecodedTransfer{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
}

// DecodeTransferFromRaw decodes a Transfer event from raw event data
func DecodeTransferFromRaw(event *RawEventMessage) (*DecodedTransfer, error) {
	if len(event.Topics) < 3 {
		return nil, errors.New("invalid log topics: expected at least 3 topics for Transfer event")
	}

	// Verify the event signature
	topic0 := event.Topics[0]
	if topic0 != TransferEventSignature.Hex() {
		return nil, fmt.Errorf("event signature mismatch: expected %s, got %s", TransferEventSignature.Hex(), topic0)
	}

	// Extract from address (topics[1])
	from := common.HexToAddress(event.Topics[1])

	// Extract to address (topics[2])
	to := common.HexToAddress(event.Topics[2])

	// Parse data as big.Int (hex string or raw bytes)
	amount := new(big.Int)
	// Remove 0x prefix if present
	data := event.Data
	if len(data) >= 2 && data[:2] == "0x" {
		data = data[2:]
	}
	amount.SetString(data, 16)

	return &DecodedTransfer{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
}

// IsTransferEvent checks if the given topics represent a Transfer event
func IsTransferEvent(topics []string) bool {
	if len(topics) < 3 {
		return false
	}
	return topics[0] == TransferEventSignature.Hex()
}
