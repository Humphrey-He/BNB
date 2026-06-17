package parser

import (
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

// isDeposit checks if a parsed event represents a valid deposit
// A deposit is valid when:
// 1. The "to" address is a watched address
// 2. The token is supported and active
func (p *Parser) isDeposit(parsed *ParsedEvent) bool {
	// Check if the to_address is a watched address
	watched, err := p.watchedAddrRepo.GetByChainIDAndAddress(
		parsed.ChainID,
		parsed.To,
	)
	if err != nil {
		p.logger.Debug("Address not found in watched addresses",
			"chain_id", parsed.ChainID,
			"address", parsed.To,
			"error", err,
		)
		return false
	}

	if !watched.IsActive {
		p.logger.Debug("Address is not active",
			"chain_id", parsed.ChainID,
			"address", parsed.To,
		)
		return false
	}

	// Check if the token is supported
	token, err := p.tokenRepo.GetByChainIDAndContract(
		parsed.ChainID,
		parsed.TokenAddress,
	)
	if err != nil {
		p.logger.Debug("Token not found",
			"chain_id", parsed.ChainID,
			"contract_address", parsed.TokenAddress,
			"error", err,
		)
		return false
	}

	if !token.IsActive {
		p.logger.Debug("Token is not active",
			"chain_id", parsed.ChainID,
			"contract_address", parsed.TokenAddress,
		)
		return false
	}

	p.logger.Info("Deposit validation passed",
		"chain_id", parsed.ChainID,
		"address", parsed.To,
		"token", parsed.TokenAddress,
		"watched_address_id", watched.ID,
		"token_id", token.ID,
	)

	return true
}

// getTokenID retrieves the token ID for a given chain and contract address
func (p *Parser) getTokenID(chainID int64, contractAddress string) (int64, error) {
	token, err := p.tokenRepo.GetByChainIDAndContract(chainID, contractAddress)
	if err != nil {
		return 0, err
	}
	return token.ID, nil
}

// buildDeposit creates a Deposit struct from a ParsedEvent
func (p *Parser) buildDeposit(parsed *ParsedEvent, idempotencyKey string) (*repository.Deposit, error) {
	tokenID, err := p.getTokenID(parsed.ChainID, parsed.TokenAddress)
	if err != nil {
		return nil, err
	}

	// Get chain config for target confirmations
	chain, err := p.chainRepo.GetByChainID(parsed.ChainID)
	if err != nil {
		return nil, err
	}

	return &repository.Deposit{
		ChainID:              parsed.ChainID,
		TokenID:              tokenID,
		TxHash:               parsed.TxHash,
		LogIndex:             int(parsed.LogIndex),
		FromAddress:          parsed.From,
		ToAddress:            parsed.To,
		Amount:               parsed.Amount,
		BlockNumber:          parsed.BlockNumber,
		Status:               repository.DepositStatusDetected,
		TargetConfirmations:  chain.FinalityConfirmations,
		IdempotencyKey:       idempotencyKey,
	}, nil
}

// saveDeposit saves a deposit to the repository with idempotency check
func (p *Parser) saveDeposit(parsed *ParsedEvent) error {
	idempotencyKey := MakeIdempotencyKey(
		parsed.ChainID,
		parsed.TxHash,
		parsed.LogIndex,
	)

	// Check if already exists
	existing, err := p.depositRepo.GetByIdempotencyKey(idempotencyKey)
	if err == nil && existing != nil {
		p.logger.Debug("Deposit already exists, skipping",
			"idempotency_key", idempotencyKey,
		)
		return nil
	}

	// Build and create the deposit
	deposit, err := p.buildDeposit(parsed, idempotencyKey)
	if err != nil {
		p.logger.Error("Failed to build deposit",
			"error", err,
			"chain_id", parsed.ChainID,
			"tx_hash", parsed.TxHash,
			"log_index", parsed.LogIndex,
			"token_address", parsed.TokenAddress,
			"to_address", parsed.To,
		)
		return err
	}

	if err := p.depositRepo.Create(deposit); err != nil {
		p.logger.Error("Failed to create deposit",
			"error", err,
			"chain_id", parsed.ChainID,
			"tx_hash", parsed.TxHash,
		)
		return err
	}

	p.logger.Info("Deposit saved successfully",
		"chain_id", parsed.ChainID,
		"tx_hash", parsed.TxHash,
		"log_index", parsed.LogIndex,
		"idempotency_key", idempotencyKey,
	)

	return nil
}
