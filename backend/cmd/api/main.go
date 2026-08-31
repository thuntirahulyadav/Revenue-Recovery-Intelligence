package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"razorpay-recovery-intelligence/backend/internal/audit"
	"razorpay-recovery-intelligence/backend/internal/config"
	"razorpay-recovery-intelligence/backend/internal/handler"
	"razorpay-recovery-intelligence/backend/internal/kafka"
	"razorpay-recovery-intelligence/backend/internal/middleware"
	"razorpay-recovery-intelligence/backend/internal/policy"
	"razorpay-recovery-intelligence/backend/internal/redis"
	"razorpay-recovery-intelligence/backend/internal/repository"
	"razorpay-recovery-intelligence/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("[RRI Backend] Starting Razorpay Recovery Intelligence Gateway in %s mode...", cfg.Env)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. Initialize PostgreSQL
	var db *repository.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = repository.NewDB(cfg.PostgresDSN())
		if err == nil {
			break
		}
		log.Printf("[PostgreSQL] Connection attempt %d/5 failed: %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("[PostgreSQL] Connection failed: %v. Persistent recovery operations require PostgreSQL.", err)
	} else {
		defer db.Close()
		// Auto seed if empty
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		_ = repository.SeedDatabaseIfEmpty(ctx, db)
		cancel()
	}

	// 2. Initialize Redis
	redisClient := redis.NewClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	defer redisClient.Close()

	// 3. Initialize Kafka / EventBus
	eventBus := kafka.NewEventBus(cfg.KafkaBrokers, cfg.KafkaEnabled, cfg.KafkaGroupID)
	defer eventBus.Close()

	// 4. Initialize Repositories
	paymentRepo := repository.NewPaymentRepo(db)
	customerRepo := repository.NewCustomerRepo(db)
	recoveryRepo := repository.NewRecoveryRepo(db)
	merchantRepo := repository.NewMerchantRepo(db)
	auditRepo := repository.NewAuditRepo(db)

	// 5. Initialize Services
	auditLogger := audit.NewLogger(auditRepo)
	policyEngine := policy.NewEngine()
	mlClient := service.NewMLClient(cfg.MLServiceURL, cfg.MLServiceTimeoutSeconds)
	recoveryService := service.NewRecoveryService(
		paymentRepo, customerRepo, recoveryRepo, merchantRepo, mlClient, policyEngine, auditLogger, eventBus, redisClient,
	)
	simulationService := service.NewSimulationService(paymentRepo)

	// 6. Initialize Handlers
	recoveryHandler := handler.NewRecoveryHandler(recoveryService, paymentRepo, customerRepo, recoveryRepo, eventBus)
	dashboardHandler := handler.NewDashboardHandler(recoveryRepo)
	simulationHandler := handler.NewSimulationHandler(simulationService)
	settingsHandler := handler.NewSettingsHandler(db)

	// 7. Setup Router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(120, time.Minute)) // 120 reqs/min

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "rri-backend-gateway",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// API v1 Routes
	v1 := router.Group("/api/v1")
	{
		// Event Ingestion
		events := v1.Group("/events")
		events.Use(middleware.IdempotencyMiddleware(redisClient))
		{
			events.POST("/payment-failed", recoveryHandler.IngestPaymentFailed)
		}

		// Payment Recovery Analysis & Details
		payments := v1.Group("/payments")
		{
			payments.GET("/:payment_id/recovery", recoveryHandler.GetPaymentRecovery)
		}

		// Recovery Actions & Queue
		recovery := v1.Group("/recovery")
		{
			recovery.GET("/opportunities", recoveryHandler.GetOpportunities)
			recovery.POST("/:payment_id/approve", recoveryHandler.ApproveRecovery)
			recovery.POST("/:payment_id/execute", middleware.IdempotencyMiddleware(redisClient), recoveryHandler.ExecuteRecovery)
		}

		// Dashboard Overview & KPIs
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", dashboardHandler.GetOverview)
		}

		// Simulation Lab (Baseline vs AI)
		sim := v1.Group("/simulation")
		{
			sim.GET("/compare", simulationHandler.CompareSimulation)
		}

		// Merchant Settings & Rules
		settings := v1.Group("/settings")
		{
			settings.GET("", settingsHandler.GetSettings)
			settings.PUT("", settingsHandler.UpdateSettings)
		}
	}

	serverAddr := ":" + cfg.Port
	log.Printf("[RRI Backend] API Gateway listening on http://0.0.0.0%s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
