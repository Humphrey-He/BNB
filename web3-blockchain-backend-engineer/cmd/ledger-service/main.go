package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/ledgerservice"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("ledger-service")

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

	service := ledgerservice.NewLedgerService(
		db,
		nc,
		repository.NewLedgerEntryRepository(db),
		repository.NewBalanceRepository(db),
		repository.NewDepositRepository(db),
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Start(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
