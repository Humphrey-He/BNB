package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/parser"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("parser")

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

	service := parser.NewParser(
		nc,
		repository.NewWatchedAddressRepository(db),
		repository.NewTokenRepository(db),
		repository.NewChainEventRepository(db),
		repository.NewDepositRepository(db),
		repository.NewChainRepository(db),
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = service.Start(ctx)
	if err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
