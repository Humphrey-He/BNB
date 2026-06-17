package handlers

import (
	"database/sql"
	"log/slog"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// Dependencies contains the repositories and infrastructure used by HTTP handlers.
type Dependencies struct {
	DB             *sql.DB
	ChainRepo      repository.ChainRepository
	TokenRepo      repository.TokenRepository
	BalanceRepo    repository.BalanceRepository
	DepositRepo    repository.DepositRepository
	WithdrawalRepo repository.WithdrawalRepository
	CheckpointRepo repository.ScanCheckpointRepository
	NATS           *nats.Conn
	Logger         *slog.Logger
}

var apiDeps Dependencies

// Configure wires the package-level handler dependencies.
func Configure(deps Dependencies) {
	apiDeps = deps
}

func handlerLogger() *slog.Logger {
	if apiDeps.Logger != nil {
		return apiDeps.Logger
	}
	return slog.Default()
}
