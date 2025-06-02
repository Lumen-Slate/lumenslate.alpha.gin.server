package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"lumenslate/internal/firebase"
	"lumenslate/internal/routes"
	"lumenslate/internal/routes/questions"

	_ "lumenslate/internal/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           ParableMind API
// @version         1.0
// @description     Backend API for managing assignments, questions, classrooms and more.
// @host            localhost:8080
// @BasePath        /

func main() {
	log.Println("🟡 Warming up server...")

	// Attempt to load secret from Cloud Run secret mount
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

	// Debug log to confirm env vars are loaded
	log.Printf("✅ FIREBASE_PROJECT_ID: %s", os.Getenv("FIREBASE_PROJECT_ID"))
	log.Printf("✅ PORT: %s", os.Getenv("PORT"))

	// Initialize Firebase
	firebase.InitFirestore()

	// Setup Gin
	router := gin.Default()
	router.Use(cors.Default())

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
	address := ":" + port

	// Log endpoints
	fmt.Printf("✅ Server running at:       http://localhost:%s\n", port)
	fmt.Printf("📘 Swagger docs available:  http://localhost:%s/docs/index.html\n", port)

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
	routes.RegisterPostRoutes(router)
	routes.RegisterQuestionBankRoutes(router)
	routes.RegisterStudentRoutes(router)
	routes.RegisterSubmissionRoutes(router)
	routes.RegisterTeacherRoutes(router)
	routes.RegisterVariableRoutes(router)
	routes.RegisterExportRoutes(router)

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
	log.Println("✅ Server exited cleanly")
}
