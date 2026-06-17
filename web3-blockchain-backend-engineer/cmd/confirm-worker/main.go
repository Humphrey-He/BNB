package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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

	rpcURL := os.Getenv("SCAN_RPC_URL")
	if rpcURL == "" {
		log.Fatal("SCAN_RPC_URL is required")
	}

	rpcClient, err := scanner.NewRPCClient(rpcURL)
	if err != nil {
		log.Fatal(err)
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
