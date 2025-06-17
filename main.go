package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
		content, _ := io.ReadAll(file)
		log.Println("📄 ENV_FILE loaded from /secrets:\n" + string(content))

		if err := godotenv.Load("/secrets/ENV_FILE"); err != nil {
			log.Println("❌ Failed to load /secrets/ENV_FILE:", err)
		} else {
			log.Println("✅ Environment loaded from /secrets/ENV_FILE")
		}
	} else {
		// Fallback to local development .env file
		log.Println("⚠️  /secrets/ENV_FILE not found, trying local .env")

		if err := godotenv.Load(); err != nil {
			log.Println("❌ Failed to load local .env:", err)
		} else {
			log.Println("✅ Environment loaded from local .env")
		}
	}

	uri := os.Getenv("MONGO_URI")
	if err := db.InitMongoDB(uri); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
}

func main() {
	log.Println("🟡 Warming up server...")

	// Set Gin log mode based on environment
	if os.Getenv("GO_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Debug log to confirm env vars are loaded
	if uri := os.Getenv("MONGO_URI"); strings.Contains(uri, "appName=") {
		appName := strings.Split(strings.Split(uri, "appName=")[1], "&")[0]
		log.Printf("✅ Mongo App Name: %s", appName)
	} else {
		log.Println("⚠️ Mongo URI does not contain appName parameter.")
	}
	log.Printf("✅ PORT: %s", os.Getenv("PORT"))

	// Setup Gin
	router := gin.Default()
	router.Use(cors.Default())

	// Configure Gin to handle trailing slashes
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Static("/media", "./media")

	// Register all routes
	registerRoutes(router)

	// Swagger & health check
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := "0.0.0.0:" + port // ✅ REQUIRED for Cloud Run

	// Log endpoints
	fmt.Printf("✅ Server running on %s\n", address)
	fmt.Printf("📘 Swagger docs available at /docs/index.html\n")

	// Run the server in a goroutine for graceful shutdown
	go func() {
		if err := router.Run(address); err != nil {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Graceful shutdown
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
	routes.SetupReportCardRoutes(router)
	routes.SetupAgentReportCardRoutes(router)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response from metadata server: %v", err)
		return
	}

	log.Printf("🔐 Cloud Run is using service account: %s", string(body))
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
