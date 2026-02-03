package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce.com/internal/app"
	"e-commerce.com/internal/config"
	"e-commerce.com/internal/db"
	"e-commerce.com/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	start := time.Now()

	if err := config.Init(); err != nil {
		log.Printf("%v", err)
	}
	if err := db.InitializeDatabase(); err != nil {
		log.Printf("db conn err : %v", err)
	}
	db.SetupGracefulShutdown()

	app, err := app.New()
	if err != nil {
		log.Fatalf("Application initialization failed: %v", err)
	}
	defer app.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-App-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	routes.SetUpRoutes(router, app)
	log.Println("🌐 Starting HTTP server on :8080...")
	log.Println("📡 Server endpoints are now available")
	log.Println("App ready in", time.Since(start).Seconds(), "seconds.")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

}
