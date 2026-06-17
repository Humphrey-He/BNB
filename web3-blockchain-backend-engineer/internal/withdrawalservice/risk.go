package withdrawalservice

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

type RiskConfig struct {
	HotWalletAddress      string
	MaxAutoApproveAmount  *big.Int
	RequireWhitelist      bool
	AllowedDestinationSet map[string]struct{}
}

func LoadRiskConfig() RiskConfig {
	return RiskConfig{
		HotWalletAddress:      normalizeAddress(os.Getenv("WITHDRAWAL_HOT_WALLET_ADDRESS")),
		MaxAutoApproveAmount:  parseBigIntEnv("WITHDRAWAL_MAX_AUTO_APPROVE_AMOUNT", "1000000000000000000"),
		RequireWhitelist:      parseBoolEnv("WITHDRAWAL_REQUIRE_WHITELIST", true),
		AllowedDestinationSet: parseAddressSet(os.Getenv("WITHDRAWAL_ALLOWED_DESTINATIONS")),
	}
}

type riskDecision struct {
	Status repository.WithdrawalStatus
	Reason string
}

func (w *WithdrawalWorker) evaluateRisk(withdrawal *repository.Withdrawal) riskDecision {
	if !isValidAddress(withdrawal.ToAddress) {
		return riskDecision{
			Status: repository.WithdrawalStatusFailed,
			Reason: fmt.Sprintf("invalid destination address: %s", withdrawal.ToAddress),
		}
	}

	if strings.TrimSpace(withdrawal.Amount) == "" || strings.TrimSpace(withdrawal.Amount) == "0" {
		return riskDecision{
			Status: repository.WithdrawalStatusFailed,
			Reason: fmt.Sprintf("invalid amount: %s", withdrawal.Amount),
		}
	}

	if w.riskConfig.HotWalletAddress != "" && normalizeAddress(withdrawal.FromAddress) != w.riskConfig.HotWalletAddress {
		return riskDecision{
			Status: repository.WithdrawalStatusManualReview,
			Reason: "withdrawal source address does not match configured hot wallet",
		}
	}

	if w.riskConfig.RequireWhitelist {
		if _, ok := w.riskConfig.AllowedDestinationSet[normalizeAddress(withdrawal.ToAddress)]; !ok {
			return riskDecision{
				Status: repository.WithdrawalStatusManualReview,
				Reason: "destination address is not in withdrawal whitelist",
			}
		}
	}

	amount, ok := new(big.Int).SetString(strings.TrimSpace(withdrawal.Amount), 10)
	if !ok {
		return riskDecision{
			Status: repository.WithdrawalStatusFailed,
			Reason: "amount is not a valid integer",
		}
	}

	if w.riskConfig.MaxAutoApproveAmount != nil && amount.Cmp(w.riskConfig.MaxAutoApproveAmount) > 0 {
		return riskDecision{
			Status: repository.WithdrawalStatusManualReview,
			Reason: "amount exceeds auto-approval threshold",
		}
	}

	balance, err := w.balanceRepo.GetByAccountChainAndToken(withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID)
	if err != nil {
		return riskDecision{
			Status: repository.WithdrawalStatusManualReview,
			Reason: "unable to load available balance for withdrawal source",
		}
	}

	available, ok := new(big.Int).SetString(strings.TrimSpace(balance.AvailableBalance), 10)
	if !ok {
		return riskDecision{
			Status: repository.WithdrawalStatusManualReview,
			Reason: "available balance is not a valid integer",
		}
	}
	if available.Cmp(amount) < 0 {
		return riskDecision{
			Status: repository.WithdrawalStatusFailed,
			Reason: "insufficient available balance",
		}
	}

	return riskDecision{Status: repository.WithdrawalStatusApproved}
}

func normalizeAddress(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

func parseAddressSet(raw string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, value := range strings.Split(raw, ",") {
		normalized := normalizeAddress(value)
		if normalized == "" {
			continue
		}
		result[normalized] = struct{}{}
	}
	return result
}

func parseBigIntEnv(key, fallback string) *big.Int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		value = fallback
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok {
		parsed, _ = new(big.Int).SetString(fallback, 10)
	}
	return parsed
}

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
