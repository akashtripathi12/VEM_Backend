package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/handlers"
	"github.com/akashtripathi12/TBO_Backend/internal/middleware"
	"github.com/akashtripathi12/TBO_Backend/internal/routes"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Load Configuration
	appConfig := config.Load()
	log.Printf("🚀 Starting TBO Backend [env: %s]", appConfig.Env)

	// 2. Initialize Store
	db := store.NewMockStore()
	log.Println("✅ Database store initialized")

	// 3. Initialize Repository/Handlers
	repo := handlers.NewRepository(appConfig, db)
	log.Println("✅ Repository initialized")

	// 4. Create Fiber App
	app := fiber.New(fiber.Config{
		AppName:               "TBO Backend API",
		ReadTimeout:           appConfig.ReadTimeout,
		WriteTimeout:          appConfig.WriteTimeout,
		BodyLimit:             appConfig.BodyLimit,
		DisableStartupMessage: false,
		ErrorHandler:          utils.GlobalErrorHandler,
	})

	// 5. Register Global Middleware (order matters!)
	app.Use(middleware.SetupRecovery())   // Panic recovery
	if appConfig.EnableLogger {
		app.Use(middleware.SetupLogger()) // Request logging
	}
	app.Use(middleware.SetupCORS(appConfig.AllowedOrigins)) // CORS

	log.Println("✅ Middleware stack configured")

	// 6. Setup Routes
	routes.SetupRoutes(app, appConfig, repo)
	log.Println("✅ Routes registered")

	// 7. Start Server with Graceful Shutdown
	go func() {
		log.Printf("🌐 Server listening on %s", appConfig.Port)
		if err := app.Listen(appConfig.Port); err != nil {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// 8. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// 9. Graceful shutdown with timeout
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
