package broadcaster

import (
	"context"
	"fmt"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

type nonceRepositoryAdapter struct {
	repo repository.NonceAllocationRepository
	ttl  time.Duration
}

func NewNonceRepositoryAdapter(repo repository.NonceAllocationRepository) NonceRepository {
	return &nonceRepositoryAdapter{
		repo: repo,
		ttl:  10 * time.Minute,
	}
}

func (a *nonceRepositoryAdapter) Allocate(ctx context.Context, chainID int64, address string) (int64, error) {
	_ = ctx
	if a.repo == nil {
		return 0, fmt.Errorf("nonce allocation repository is not configured")
	}

	nonce, err := a.repo.GetNextAvailableNonce(chainID, address)
	if err != nil {
		return 0, err
	}

	err = a.repo.Create(&repository.NonceAllocation{
		ChainID:     chainID,
		FromAddress: address,
		Nonce:       nonce,
		Status:      repository.NonceStatusAllocated,
		ExpiresAt:   time.Now().Add(a.ttl),
	})
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (a *nonceRepositoryAdapter) Release(ctx context.Context, chainID int64, address string, nonce int64) error {
	_ = ctx
	if a.repo == nil {
		return fmt.Errorf("nonce allocation repository is not configured")
	}
	return a.repo.MarkExpired(chainID, address, nonce)
}
