package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"lumenslate/internal/db"
	"lumenslate/internal/routes"
	"lumenslate/internal/routes/questions"
	"lumenslate/internal/service"
	"lumenslate/tasks"

	_ "lumenslate/internal/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Lumen Slate API
// @version         1.0
// @description     Backend API for managing assignments, questions, classrooms and more.
// @host            localhost:8080
// @BasePath        /

func init() {
	logADCIdentity()

	if file, err := os.Open("/secrets/ENV_FILE"); err == nil {
		defer file.Close()
		_, _ = io.ReadAll(file)
		_ = godotenv.Load("/secrets/ENV_FILE")
	} else {
		_ = godotenv.Load()
	}

	uri := os.Getenv("MONGO_URI")
	if err := db.InitMongoDB(uri); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
}

func main() {
	startTime := time.Now()

	gin.SetMode(os.Getenv("GIN_MODE")) // This will suppress the debug logs
	gin.DisableConsoleColor()
	router := gin.New()

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"}, // Skip logging health checks
	}))
	router.Use(gin.Recovery())
	router.Use(cors.Default())
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Static("/media", "./media")

	// Initialize metrics collector for monitoring
	metricsCollector := initializeMetricsCollector()

	// Create API v1 group
	apiV1 := router.Group("/api/v1")

	// Register all API routes under /api/v1
	registerRoutes(apiV1, metricsCollector, startTime)

	// Health and docs endpoints remain at root
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize and start Asynq server for background task processing
	asynqServer := initializeAsynqServer()
	if err := asynqServer.Start(); err != nil {
		log.Fatalf("❌ Failed to start Asynq server: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := "0.0.0.0:" + port // ✅ REQUIRED for Cloud Run
	log.Printf("[BOOT] Starting server on %s", address)

	go func() {
		if err := router.Run(address); err != nil {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	gracefulShutdown(asynqServer, metricsCollector)
}

// Change router type from *gin.Engine to gin.IRoutes to allow both *gin.Engine and *gin.RouterGroup
func registerRoutes(router *gin.RouterGroup, metricsCollector *service.MetricsCollector, startTime time.Time) {
	routes.RegisterAssignmentRoutes(router)
	routes.RegisterClassroomRoutes(router)
	routes.RegisterCommentRoutes(router)
	routes.RegisterThreadRoutes(router)
	routes.RegisterQuestionBankRoutes(router)
	routes.RegisterStudentRoutes(router)
	routes.RegisterSubmissionRoutes(router)
	routes.RegisterTeacherRoutes(router)
	routes.RegisterVariableRoutes(router)
	routes.RegisterAIRoutes(router)
	routes.SetupSubjectReportRoutes(router)
	routes.SetupReportCardRoutes(router)
	routes.SetupAgentReportCardRoutes(router)
	routes.RegisterUserRoutes(router)
	questions.RegisterMCQRoutes(router)
	questions.RegisterMSQRoutes(router)
	questions.RegisterNATRoutes(router)
	questions.RegisterSubjectiveRoutes(router)
}

func logADCIdentity() {
	metadataURL := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email"

	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		log.Printf("❌ Failed to create request to metadata server: %v", err)
		return
	}
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to call metadata server: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Println("[BOOT] Successfully called metadata server for ADC identity")
}

// initializeAsynqServer creates and configures the Asynq server with task handlers
func initializeAsynqServer() *service.AsynqServer {
	// Get Redis configuration from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Create Asynq server with default concurrency
	asynqServer := service.NewAsynqServer(redisAddr, 0) // 0 uses default from env or 10

	// Register document processing task handler
	if err := asynqServer.RegisterTaskHandler(tasks.TypeAddDocumentToCorpus, tasks.HandleAddDocumentToCorpusTask); err != nil {
		log.Fatalf("❌ Failed to register document task handler: %v", err)
	}

	log.Printf("[BOOT] Asynq server initialized with Redis at %s", redisAddr)
	return asynqServer
}

// initializeMetricsCollector creates and configures the metrics collector
func initializeMetricsCollector() *service.MetricsCollector {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	metricsCollector, err := service.NewMetricsCollector(redisAddr)
	if err != nil {
		log.Fatalf("❌ Failed to initialize metrics collector: %v", err)
	}

	// Set the global metrics collector for task handlers
	tasks.SetMetricsCollector(metricsCollector)

	log.Printf("[BOOT] Metrics collector initialized with Redis at %s", redisAddr)
	return metricsCollector
}

func gracefulShutdown(asynqServer *service.AsynqServer, metricsCollector *service.MetricsCollector) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Stop Asynq server first
	if asynqServer != nil {
		asynqServer.Stop()
	}

	// Close metrics collector
	if metricsCollector != nil {
		if err := metricsCollector.Close(); err != nil {
			log.Printf("❌ Error closing metrics collector: %v", err)
		}
	}

	// Close MongoDB connection
	if err := db.CloseMongoDB(); err != nil {
		log.Printf("❌ Error closing MongoDB connection: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}
