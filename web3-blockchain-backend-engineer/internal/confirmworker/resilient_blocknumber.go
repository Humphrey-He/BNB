package confirmworker

import (
	"context"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/scanner"
)

type resilientBlockNumberClient struct {
	client scanner.Client
}

func NewResilientBlockNumberClient(chainID int64, repo repository.RPCProviderRepository) (RPCClient, error) {
	client, err := scanner.NewResilientClientFromRepo(chainID, repo)
	if err != nil {
		return nil, err
	}
	return &resilientBlockNumberClient{client: client}, nil
}

func (c *resilientBlockNumberClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}
