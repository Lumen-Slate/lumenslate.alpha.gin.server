package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"lumenslate/internal/db"
	"lumenslate/internal/routes"
	"lumenslate/internal/routes/questions"

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
	gin.SetMode(gin.DebugMode) // This will suppress the debug logs
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
	registerRoutes(router)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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

	gracefulShutdown()
}

func registerRoutes(router *gin.Engine) {
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
	routes.SetupAgentReportCardRoutes(router)
	routes.SetupAssignmentResultsRoutes(router)

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

func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	if err := db.CloseMongoDB(); err != nil {
		log.Printf("❌ Error closing MongoDB connection: %v", err)
	}
	log.Println("✅ Server exited cleanly")
}
