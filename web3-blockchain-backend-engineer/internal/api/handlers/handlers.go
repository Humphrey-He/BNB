package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/asset-platform/multi-chain-asset-platform/internal/rpcgateway"
	"github.com/asset-platform/multi-chain-asset-platform/internal/withdrawalservice"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary 健康检查
// @Description 服务健康状态检查
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// ChainResponse represents a chain object
type ChainResponse struct {
	ID                    int64  `json:"id"`
	ChainID               int64  `json:"chain_id"`
	Name                  string `json:"name"`
	NativeSymbol          string `json:"native_symbol"`
	FinalityConfirmations int    `json:"finality_confirmations"`
	IsActive              bool   `json:"is_active"`
}

// ListChains godoc
// @Summary 获取链列表
// @Description 获取所有已配置的链信息
// @Tags chains
// @Produce json
// @Success 200 {array} ChainResponse
// @Failure 500 {object} map[string]string
// @Router /chains [get]
func ListChains(c *gin.Context) {
	chains, err := apiDeps.ChainRepo.List()
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	resp := make([]ChainResponse, 0, len(chains))
	for _, chain := range chains {
		resp = append(resp, ChainResponse{
			ID:                    chain.ID,
			ChainID:               chain.ChainID,
			Name:                  chain.Name,
			NativeSymbol:          chain.NativeSymbol,
			FinalityConfirmations: chain.FinalityConfirmations,
			IsActive:              chain.IsActive,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// TokenResponse represents a token object
type TokenResponse struct {
	ID              int64  `json:"id"`
	ChainID         int64  `json:"chain_id"`
	ContractAddress string `json:"contract_address"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	IsNative        bool   `json:"is_native"`
	IsActive        bool   `json:"is_active"`
}

// ListTokens godoc
// @Summary 获取Token列表
// @Description 获取所有已配置的Token信息
// @Tags tokens
// @Produce json
// @Param chain_id query int false "Chain ID to filter by"
// @Success 200 {array} TokenResponse
// @Failure 500 {object} map[string]string
// @Router /tokens [get]
func ListTokens(c *gin.Context) {
	var (
		tokens []*repository.Token
		err    error
	)

	chainID, hasChainID, err := parseOptionalInt64(c, "chain_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chain_id"})
		return
	}

	if hasChainID {
		tokens, err = apiDeps.TokenRepo.ListByChainID(chainID)
	} else {
		tokens, err = apiDeps.TokenRepo.List()
	}
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	resp := make([]TokenResponse, 0, len(tokens))
	for _, token := range tokens {
		resp = append(resp, TokenResponse{
			ID:              token.ID,
			ChainID:         token.ChainID,
			ContractAddress: token.ContractAddress,
			Symbol:          token.Symbol,
			Decimals:        token.Decimals,
			IsNative:        token.IsNative,
			IsActive:        token.IsActive,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// BalanceResponse represents a balance object
type BalanceResponse struct {
	AccountAddress   string `json:"account_address"`
	ChainID          int64  `json:"chain_id"`
	TokenID          int64  `json:"token_id"`
	AvailableBalance string `json:"available_balance"`
	FrozenBalance    string `json:"frozen_balance"`
}

// GetAddressBalances godoc
// @Summary 获取地址余额
// @Description 获取指定地址在所有链和Token上的余额
// @Tags balances
// @Produce json
// @Param address path string true "Wallet address"
// @Param chain_id query int false "Filter by chain ID"
// @Success 200 {array} BalanceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /addresses/{address}/balances [get]
func GetAddressBalances(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}

	balances, err := apiDeps.BalanceRepo.ListByAccountAddress(address)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	chainID, hasChainID, err := parseOptionalInt64(c, "chain_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chain_id"})
		return
	}

	resp := make([]BalanceResponse, 0, len(balances))
	for _, balance := range balances {
		if hasChainID && balance.ChainID != chainID {
			continue
		}
		resp = append(resp, BalanceResponse{
			AccountAddress:   balance.AccountAddress,
			ChainID:          balance.ChainID,
			TokenID:          balance.TokenID,
			AvailableBalance: balance.AvailableBalance,
			FrozenBalance:    balance.FrozenBalance,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// DepositResponse represents a deposit object
type DepositResponse struct {
	ID            int64  `json:"id"`
	ChainID       int64  `json:"chain_id"`
	TokenID       int64  `json:"token_id"`
	TxHash        string `json:"tx_hash"`
	FromAddress   string `json:"from_address"`
	ToAddress     string `json:"to_address"`
	Amount        string `json:"amount"`
	BlockNumber   int64  `json:"block_number"`
	Status        string `json:"status"`
	Confirmations int    `json:"confirmations"`
	CreatedAt     string `json:"created_at"`
}

// ListDeposits godoc
// @Summary 获取充值记录列表
// @Description 获取充值记录，支持分页和过滤
// @Tags deposits
// @Produce json
// @Param address query string false "Filter by to_address"
// @Param chain_id query int false "Filter by chain ID"
// @Param status query string false "Filter by status (detected, pending_confirmation, confirmed, orphaned)"
// @Param limit query int false "Limit results (default 100)"
// @Success 200 {array} DepositResponse
// @Failure 500 {object} map[string]string
// @Router /deposits [get]
func ListDeposits(c *gin.Context) {
	limit, err := parseLimit(c, 100, 500)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	chainID, hasChainID, err := parseOptionalInt64(c, "chain_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chain_id"})
		return
	}
	address := strings.TrimSpace(c.Query("address"))
	status := strings.TrimSpace(c.Query("status"))

	var deposits []*repository.Deposit
	switch {
	case address != "" && hasChainID:
		deposits, err = apiDeps.DepositRepo.ListByAddress(chainID, address, limit)
	case status != "":
		deposits, err = apiDeps.DepositRepo.ListByStatus(repository.DepositStatus(status), limit)
	case hasChainID:
		deposits, err = apiDeps.DepositRepo.ListByChainID(chainID, limit)
	default:
		deposits, err = apiDeps.DepositRepo.List(limit)
	}
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	resp := make([]DepositResponse, 0, len(deposits))
	for _, deposit := range deposits {
		if address != "" && !strings.EqualFold(deposit.ToAddress, address) {
			continue
		}
		if hasChainID && deposit.ChainID != chainID {
			continue
		}
		if status != "" && string(deposit.Status) != status {
			continue
		}
		resp = append(resp, DepositResponse{
			ID:            deposit.ID,
			ChainID:       deposit.ChainID,
			TokenID:       deposit.TokenID,
			TxHash:        deposit.TxHash,
			FromAddress:   deposit.FromAddress,
			ToAddress:     deposit.ToAddress,
			Amount:        deposit.Amount,
			BlockNumber:   deposit.BlockNumber,
			Status:        string(deposit.Status),
			Confirmations: deposit.Confirmations,
			CreatedAt:     formatTime(deposit.CreatedAt),
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GetDeposit godoc
// @Summary 获取充值详情
// @Description 根据ID获取充值详情
// @Tags deposits
// @Produce json
// @Param id path int true "Deposit ID"
// @Success 200 {object} DepositResponse
// @Failure 404 {object} map[string]string
// @Router /deposits/{id} [get]
func GetDeposit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	deposit, err := apiDeps.DepositRepo.GetByID(id)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, DepositResponse{
		ID:            deposit.ID,
		ChainID:       deposit.ChainID,
		TokenID:       deposit.TokenID,
		TxHash:        deposit.TxHash,
		FromAddress:   deposit.FromAddress,
		ToAddress:     deposit.ToAddress,
		Amount:        deposit.Amount,
		BlockNumber:   deposit.BlockNumber,
		Status:        string(deposit.Status),
		Confirmations: deposit.Confirmations,
		CreatedAt:     formatTime(deposit.CreatedAt),
	})
}

// CreateWithdrawalRequest represents withdrawal creation request
type CreateWithdrawalRequest struct {
	ChainID   int64  `json:"chain_id" binding:"required"`
	TokenID   int64  `json:"token_id" binding:"required"`
	ToAddress string `json:"to_address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
}

// WithdrawalResponse represents a withdrawal object
type WithdrawalResponse struct {
	ID          int64  `json:"id"`
	ChainID     int64  `json:"chain_id"`
	TokenID     int64  `json:"token_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
	TxHash      string `json:"tx_hash,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type ReviewWithdrawalRequest struct {
	Reason string `json:"reason"`
}

// CreateWithdrawal godoc
// @Summary 创建提现申请
// @Description 创建新的提现申请，会冻结对应余额
// @Tags withdrawals
// @Accept json
// @Produce json
// @Param request body CreateWithdrawalRequest true "Withdrawal request"
// @Param Idempotency-Key header string false "Idempotency key for preventing duplicate requests"
// @Success 201 {object} WithdrawalResponse
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /withdrawals [post]
func CreateWithdrawal(c *gin.Context) {
	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	fromAddress := strings.TrimSpace(c.GetHeader("X-From-Address"))
	if fromAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-From-Address header is required"})
		return
	}

	existing, err := apiDeps.WithdrawalRepo.GetByIdempotencyKey(idempotencyKey)
	if err == nil && existing != nil {
		c.JSON(http.StatusOK, toWithdrawalResponse(existing))
		return
	}
	if err != nil && err != sql.ErrNoRows {
		respondRepositoryError(c, err)
		return
	}

	withdrawal := &repository.Withdrawal{
		ChainID:        req.ChainID,
		TokenID:        req.TokenID,
		FromAddress:    fromAddress,
		ToAddress:      strings.TrimSpace(req.ToAddress),
		Amount:         strings.TrimSpace(req.Amount),
		Status:         repository.WithdrawalStatusCreated,
		IdempotencyKey: idempotencyKey,
	}

	if !common.IsHexAddress(withdrawal.FromAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-From-Address must be a valid hex address"})
		return
	}
	if !common.IsHexAddress(withdrawal.ToAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_address must be a valid hex address"})
		return
	}
	if strings.EqualFold(withdrawal.FromAddress, withdrawal.ToAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to addresses must differ"})
		return
	}
	if _, err := parsePositiveIntegerString(withdrawal.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be a positive base-10 integer"})
		return
	}

	chain, err := apiDeps.ChainRepo.GetByChainID(withdrawal.ChainID)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}
	if !chain.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain is inactive"})
		return
	}

	token, err := apiDeps.TokenRepo.GetByID(withdrawal.TokenID)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}
	if token.ChainID != chain.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token does not belong to chain_id"})
		return
	}
	if !token.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is inactive"})
		return
	}

	if apiDeps.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "withdrawal service is not configured"})
		return
	}

	service := newWithdrawalCreationService(apiDeps.DB)
	if err := service.Create(c.Request.Context(), withdrawal); err != nil {
		if strings.Contains(err.Error(), "insufficient available balance") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		handlerLogger().Error("failed to create withdrawal transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create withdrawal"})
		return
	}

	if apiDeps.NATS != nil {
		event := withdrawalservice.WithdrawalCreatedEvent{
			WithdrawalID:   withdrawal.ID,
			ChainID:        withdrawal.ChainID,
			TokenID:        withdrawal.TokenID,
			FromAddress:    withdrawal.FromAddress,
			ToAddress:      withdrawal.ToAddress,
			Amount:         withdrawal.Amount,
			IdempotencyKey: withdrawal.IdempotencyKey,
		}

		data, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			handlerLogger().Error("failed to marshal withdrawal created event", "error", marshalErr, "withdrawal_id", withdrawal.ID)
		} else if publishErr := apiDeps.NATS.Publish(withdrawalservice.SubjectWithdrawalCreated, data); publishErr != nil {
			handlerLogger().Error("failed to publish withdrawal created event", "error", publishErr, "withdrawal_id", withdrawal.ID)
		}
	}

	c.JSON(http.StatusCreated, toWithdrawalResponse(withdrawal))
}

// ListWithdrawals godoc
// @Summary 获取提现记录列表
// @Description 获取提现记录，支持分页和过滤
// @Tags withdrawals
// @Produce json
// @Param address query string false "Filter by from_address"
// @Param chain_id query int false "Filter by chain ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results (default 100)"
// @Success 200 {array} WithdrawalResponse
// @Failure 500 {object} map[string]string
// @Router /withdrawals [get]
func ListWithdrawals(c *gin.Context) {
	limit, err := parseLimit(c, 100, 500)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	chainID, hasChainID, err := parseOptionalInt64(c, "chain_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chain_id"})
		return
	}
	address := strings.TrimSpace(c.Query("address"))
	status := strings.TrimSpace(c.Query("status"))

	var withdrawals []*repository.Withdrawal
	switch {
	case address != "" && hasChainID:
		withdrawals, err = apiDeps.WithdrawalRepo.ListByFromAddress(chainID, address, limit)
	case status != "":
		withdrawals, err = apiDeps.WithdrawalRepo.ListByStatus(repository.WithdrawalStatus(status), limit)
	case hasChainID:
		withdrawals, err = apiDeps.WithdrawalRepo.ListByChainID(chainID, limit)
	default:
		withdrawals, err = apiDeps.WithdrawalRepo.List(limit)
	}
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	resp := make([]WithdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		if address != "" && !strings.EqualFold(withdrawal.FromAddress, address) {
			continue
		}
		if hasChainID && withdrawal.ChainID != chainID {
			continue
		}
		if status != "" && string(withdrawal.Status) != status {
			continue
		}
		resp = append(resp, toWithdrawalResponse(withdrawal))
	}

	c.JSON(http.StatusOK, resp)
}

// GetWithdrawal godoc
// @Summary 获取提现详情
// @Description 根据ID获取提现详情
// @Tags withdrawals
// @Produce json
// @Param id path int true "Withdrawal ID"
// @Success 200 {object} WithdrawalResponse
// @Failure 404 {object} map[string]string
// @Router /withdrawals/{id} [get]
func GetWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	withdrawal, err := apiDeps.WithdrawalRepo.GetByID(id)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, toWithdrawalResponse(withdrawal))
}

func ApproveWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if apiDeps.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "withdrawal review service is not configured"})
		return
	}

	service := newWithdrawalReviewService(apiDeps.DB)
	withdrawal, err := service.Approve(c.Request.Context(), id)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "manual_review"):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			handlerLogger().Error("failed to approve withdrawal", "withdrawal_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve withdrawal"})
		}
		return
	}

	if apiDeps.NATS != nil {
		event := withdrawalservice.WithdrawalApprovedEvent{
			WithdrawalID: withdrawal.ID,
			ChainID:      withdrawal.ChainID,
			TokenID:      withdrawal.TokenID,
			FromAddress:  withdrawal.FromAddress,
			ToAddress:    withdrawal.ToAddress,
			Amount:       withdrawal.Amount,
		}
		data, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			handlerLogger().Error("failed to marshal approved withdrawal event", "withdrawal_id", withdrawal.ID, "error", marshalErr)
		} else if publishErr := apiDeps.NATS.Publish(withdrawalservice.SubjectWithdrawalApproved, data); publishErr != nil {
			handlerLogger().Error("failed to publish approved withdrawal event", "withdrawal_id", withdrawal.ID, "error", publishErr)
		}
	}

	c.JSON(http.StatusOK, toWithdrawalResponse(withdrawal))
}

func RejectWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ReviewWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "rejected by operator"
	}

	if apiDeps.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "withdrawal review service is not configured"})
		return
	}

	service := newWithdrawalReviewService(apiDeps.DB)
	withdrawal, err := service.Reject(c.Request.Context(), id, reason)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "manual_review"):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "frozen balance"):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			handlerLogger().Error("failed to reject withdrawal", "withdrawal_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject withdrawal"})
		}
		return
	}

	c.JSON(http.StatusOK, toWithdrawalResponse(withdrawal))
}

// TransactionResponse represents a transaction/history entry
type TransactionResponse struct {
	Type        string `json:"type"` // deposit or withdrawal
	ChainID     int64  `json:"chain_id"`
	TokenID     int64  `json:"token_id"`
	TxHash      string `json:"tx_hash"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
	BlockNumber int64  `json:"block_number,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// GetAddressTransactions godoc
// @Summary获取地址交易历史
// @Description 获取指定地址的所有交易记录（充值和提现）
// @Tags transactions
// @Produce json
// @Param address path string true "Wallet address"
// @Param limit query int false "Limit results (default 100)"
// @Success 200 {array} TransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /addresses/{address}/transactions [get]
func GetAddressTransactions(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}

	limit, err := parseLimit(c, 100, 500)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	var history []TransactionResponse

	deposits, err := apiDeps.DepositRepo.List(limit)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}
	for _, deposit := range deposits {
		if strings.EqualFold(deposit.ToAddress, address) {
			history = append(history, TransactionResponse{
				Type:        "deposit",
				ChainID:     deposit.ChainID,
				TokenID:     deposit.TokenID,
				TxHash:      deposit.TxHash,
				FromAddress: deposit.FromAddress,
				ToAddress:   deposit.ToAddress,
				Amount:      deposit.Amount,
				Status:      string(deposit.Status),
				BlockNumber: deposit.BlockNumber,
				CreatedAt:   formatTime(deposit.CreatedAt),
			})
		}
	}

	withdrawals, err := apiDeps.WithdrawalRepo.List(limit)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}
	for _, withdrawal := range withdrawals {
		if strings.EqualFold(withdrawal.FromAddress, address) {
			history = append(history, TransactionResponse{
				Type:        "withdrawal",
				ChainID:     withdrawal.ChainID,
				TokenID:     withdrawal.TokenID,
				TxHash:      withdrawal.TxHash,
				FromAddress: withdrawal.FromAddress,
				ToAddress:   withdrawal.ToAddress,
				Amount:      withdrawal.Amount,
				Status:      string(withdrawal.Status),
				CreatedAt:   formatTime(withdrawal.CreatedAt),
			})
		}
	}

	c.JSON(http.StatusOK, history)
}

// ChainStatusResponse represents scanner chain status
type ChainStatusResponse struct {
	ChainID          int64  `json:"chain_id"`
	Name             string `json:"name"`
	LastScannedBlock int64  `json:"last_scanned_block"`
	LatestBlock      int64  `json:"latest_block"`
	ScanLag          int64  `json:"scan_lag"`
	IsActive         bool   `json:"is_active"`
	RPCHealthy       bool   `json:"rpc_healthy"`
	RPCProvider      string `json:"rpc_provider,omitempty"`
	RPCError         string `json:"rpc_error,omitempty"`
	ProviderCount    int    `json:"provider_count"`
}

// GetChainStatus godoc
// @Summary 获取链扫描状态
// @Description 获取各链的扫描进度和延迟
// @Tags system
// @Produce json
// @Success 200 {array} ChainStatusResponse
// @Failure 500 {object} map[string]string
// @Router /chain-status [get]
func GetChainStatus(c *gin.Context) {
	chains, err := apiDeps.ChainRepo.List()
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	resp := make([]ChainStatusResponse, 0, len(chains))
	for _, chain := range chains {
		checkpoint, cpErr := apiDeps.CheckpointRepo.GetByChainID(chain.ID)
		lastScannedBlock := int64(0)
		if cpErr == nil && checkpoint != nil {
			lastScannedBlock = checkpoint.LastScannedBlock
		} else if cpErr != nil && cpErr != sql.ErrNoRows {
			handlerLogger().Warn("failed to load checkpoint", "chain_id", chain.ID, "error", cpErr)
		}

		latestBlock := lastScannedBlock
		rpcHealthy := false
		rpcProvider := ""
		rpcError := ""
		providerCount := 0
		if apiDeps.RPCProviderRepo != nil {
			client, rpcErr := rpcgateway.NewResilientClient(chain.ID, apiDeps.RPCProviderRepo, rpcgateway.DefaultCallOptions())
			if rpcErr != nil {
				rpcError = rpcErr.Error()
				handlerLogger().Warn("failed to build rpc client for chain status", "chain_id", chain.ID, "error", rpcErr)
			} else {
				snapshots := client.InspectProviders()
				providerCount = len(snapshots)
				for _, snapshot := range snapshots {
					if snapshot.IsActive && snapshot.CircuitState != "open" {
						rpcProvider = snapshot.ProviderName
						break
					}
				}

				var head uint64
				callErr := client.Call(c.Request.Context(), func(callCtx context.Context, ethClient *ethclient.Client) error {
					n, err := ethClient.BlockNumber(callCtx)
					if err != nil {
						return err
					}
					head = n
					return nil
				})
				if callErr != nil {
					rpcError = callErr.Error()
					handlerLogger().Warn("failed to fetch latest block for chain status", "chain_id", chain.ID, "error", callErr)
				} else {
					latestBlock = int64(head)
					rpcHealthy = true
				}
			}
		}

		scanLag := latestBlock - lastScannedBlock
		if scanLag < 0 {
			scanLag = 0
		}

		resp = append(resp, ChainStatusResponse{
			ChainID:          chain.ChainID,
			Name:             chain.Name,
			LastScannedBlock: lastScannedBlock,
			LatestBlock:      latestBlock,
			ScanLag:          scanLag,
			IsActive:         chain.IsActive,
			RPCHealthy:       rpcHealthy,
			RPCProvider:      rpcProvider,
			RPCError:         rpcError,
			ProviderCount:    providerCount,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func toWithdrawalResponse(withdrawal *repository.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		ID:          withdrawal.ID,
		ChainID:     withdrawal.ChainID,
		TokenID:     withdrawal.TokenID,
		FromAddress: withdrawal.FromAddress,
		ToAddress:   withdrawal.ToAddress,
		Amount:      withdrawal.Amount,
		Status:      string(withdrawal.Status),
		TxHash:      withdrawal.TxHash,
		CreatedAt:   formatTime(withdrawal.CreatedAt),
	}
}
