package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/broadcaster"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("broadcaster")

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

	nonceRepo := broadcaster.NewNonceRepositoryAdapter(repository.NewNonceAllocationRepository(db))
	rpcClient := &broadcaster.ReceiptRPCAdapter{}

	service := broadcaster.NewBroadcaster(
		db,
		nc,
		repository.NewWithdrawalRepository(db),
		nonceRepo,
		rpcClient,
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Start(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
