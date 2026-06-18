package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/confirmworker"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/scanner"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("confirm-worker")

	db, err := app.OpenPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nc, err := app.ConnectNATS(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	chainDBID, err := mustEnvInt64("SCAN_CHAIN_DB_ID")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient, err := confirmworker.NewResilientBlockNumberClient(chainDBID, repository.NewRPCProviderRepository(db))
	if err != nil {
		logger.Warn("failed to initialize resilient RPC client, falling back to SCAN_RPC_URL", "error", err)
		rpcURL := os.Getenv("SCAN_RPC_URL")
		if rpcURL == "" {
			log.Fatal("SCAN_RPC_URL is required when rpc_providers are unavailable")
		}
		rpcClient, err = scanner.NewRPCClient(rpcURL)
		if err != nil {
			log.Fatal(err)
		}
	}

	service := confirmworker.NewConfirmWorker(
		nc,
		repository.NewDepositRepository(db),
		repository.NewChainRepository(db),
		rpcClient,
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = service.Start(ctx)
	if err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

func mustEnvInt64(key string) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, os.ErrNotExist
	}
	return strconv.ParseInt(value, 10, 64)
}
