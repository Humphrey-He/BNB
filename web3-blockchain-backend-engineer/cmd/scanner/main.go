package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/scanner"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("scanner")

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
	rpcURL := os.Getenv("SCAN_RPC_URL")
	if rpcURL == "" {
		log.Fatal("SCAN_RPC_URL is required")
	}

	cfgRuntime := scanner.DefaultConfig(chainDBID)
	if batchSize := os.Getenv("SCAN_BATCH_SIZE"); batchSize != "" {
		parsed, err := strconv.ParseUint(batchSize, 10, 64)
		if err != nil {
			log.Fatalf("invalid SCAN_BATCH_SIZE: %v", err)
		}
		cfgRuntime.BatchSize = parsed
	}

	rpcClient, err := scanner.NewRPCClient(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	service := scanner.NewScanner(
		chainDBID,
		rpcClient,
		repository.NewScanCheckpointRepository(db),
		repository.NewBlockRepository(db),
		nc,
		logger,
		scanner.NewMetrics("asset_platform"),
		cfgRuntime,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil && err != context.Canceled {
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
