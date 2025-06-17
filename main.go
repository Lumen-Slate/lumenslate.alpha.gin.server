package main

import (
	"fmt"
	"io"
	"log"
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
	// Load ENV_FILE from /secrets (used in Cloud Run secrets mount)
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

	// Init DB
	uri := os.Getenv("MONGO_URI")
	if err := db.InitMongoDB(uri); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
}

func main() {
	log.Println("🟡 Starting server...")

	// Set release/debug mode based on GO_ENV
	if os.Getenv("GO_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Debug Mongo URI
	if uri := os.Getenv("MONGO_URI"); strings.Contains(uri, "appName=") {
		appName := strings.Split(strings.Split(uri, "appName=")[1], "&")[0]
		log.Printf("✅ Mongo App Name: %s", appName)
	} else {
		log.Println("⚠️ Mongo URI does not contain appName parameter.")
	}

	// Setup router
	router := gin.Default()
	router.Use(cors.Default())
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Static("/media", "./media")

	// Register routes
	registerRoutes(router)

	// Health check and Swagger
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Always run on port 8080 (Cloud Run expects this)
	const port = "8080"
	address := ":" + port

	fmt.Printf("✅ Server running on port %s\n", port)
	fmt.Printf("📘 Swagger docs available at /docs/index.html\n")

	// Start server
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
	routes.SetupReportCardRoutes(router)

	questions.RegisterMCQRoutes(router)
	questions.RegisterMSQRoutes(router)
	questions.RegisterNATRoutes(router)
	questions.RegisterSubjectiveRoutes(router)
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
