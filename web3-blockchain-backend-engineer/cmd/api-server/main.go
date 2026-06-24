package main

import (
	"log"
	"os"

	_ "github.com/asset-platform/multi-chain-asset-platform/docs" // swag
	"github.com/asset-platform/multi-chain-asset-platform/internal/api/handlers"
	"github.com/asset-platform/multi-chain-asset-platform/internal/app"
	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Multi-Chain Asset Platform API
// @version 1.0
// @description 多链资产后端平台API接口文档，支持余额查询、充值记录、提现申请、链状态等
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	cfg := app.LoadEnvConfig()
	logger := app.NewLogger("api-server")

	db, err := app.OpenPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	natsConn, err := app.ConnectNATS(cfg)
	if err != nil {
		logger.Warn("failed to connect to NATS, continuing without event publishing", "error", err)
	} else {
		defer natsConn.Close()
	}

	handlers.Configure(handlers.Dependencies{
		DB:              db,
		ChainRepo:       repository.NewChainRepository(db),
		TokenRepo:       repository.NewTokenRepository(db),
		RPCProviderRepo: repository.NewRPCProviderRepository(db),
		BalanceRepo:     repository.NewBalanceRepository(db),
		DepositRepo:     repository.NewDepositRepository(db),
		WithdrawalRepo:  repository.NewWithdrawalRepository(db),
		CheckpointRepo:  repository.NewScanCheckpointRepository(db),
		NATS:            natsConn,
		Logger:          logger,
	})

	r := gin.Default()

	// Swagger docs endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/chains", handlers.ListChains)
		v1.GET("/tokens", handlers.ListTokens)
		v1.GET("/addresses/:address/balances", handlers.GetAddressBalances)
		v1.GET("/deposits", handlers.ListDeposits)
		v1.GET("/deposits/:id", handlers.GetDeposit)
		v1.GET("/addresses/:address/transactions", handlers.GetAddressTransactions)
	}

	management := r.Group("/api/v1")
	management.Use(handlers.RequireManagementToken())
	{
		management.POST("/withdrawals", handlers.CreateWithdrawal)
		management.GET("/withdrawals", handlers.ListWithdrawals)
		management.GET("/withdrawals/:id", handlers.GetWithdrawal)
		management.POST("/withdrawals/:id/approve", handlers.ApproveWithdrawal)
		management.POST("/withdrawals/:id/reject", handlers.RejectWithdrawal)
		management.GET("/chain-status", handlers.GetChainStatus)
	}

	// Health check
	r.GET("/healthz", handlers.HealthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	log.Printf("Server starting on %s:%s", bindAddr, port)
	log.Printf("Management endpoints require Authorization: Bearer <token> via %s", "API_AUTH_TOKEN")
	log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", port)

	if err := r.Run(bindAddr + ":" + port); err != nil {
		log.Fatal(err)
	}
}
