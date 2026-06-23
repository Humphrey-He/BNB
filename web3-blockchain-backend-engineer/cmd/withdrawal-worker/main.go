package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/withdrawalservice"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("withdrawal-worker")

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

	service := withdrawalservice.NewWithdrawalWorker(
		nc,
		repository.NewWithdrawalRepository(db),
		repository.NewBalanceRepository(db),
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Start(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
