package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/scanner"
)

func main() {
	cfg := app.LoadEnvConfig()
	db, err := app.OpenPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	chainDBID, err := mustEnvInt64("SCAN_CHAIN_DB_ID")
	if err != nil {
		log.Fatal(err)
	}

	client, err := scanner.NewResilientClientFromRepo(chainDBID, repository.NewRPCProviderRepository(db))
	if err != nil {
		log.Fatal(err)
	}

	reporter, ok := client.(scanner.HealthReporter)
	if !ok {
		log.Fatal("health reporter unavailable")
	}

	head, headErr := client.BlockNumber(context.Background())
	output := struct {
		ChainDBID int64                     `json:"chain_db_id"`
		Head      uint64                    `json:"head,omitempty"`
		HeadError string                    `json:"head_error,omitempty"`
		Providers []scanner.ResilientHealth `json:"providers"`
	}{
		ChainDBID: chainDBID,
		Providers: reporter.InspectProviders(),
	}
	if headErr != nil {
		output.HeadError = headErr.Error()
	} else {
		output.Head = head
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
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
