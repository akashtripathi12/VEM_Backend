package main

import (
	"log"
	"os"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/handlers"
	"github.com/akashtripathi12/TBO_Backend/internal/routes"
	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Load Config
	cfg := config.Load()

	// Initialize Store
	store.InitDB()

	app := fiber.New()

	// Initialize Repository
	repo := handlers.NewRepository(cfg)

	// Setup Routes
	routes.SetupRoutes(app, cfg, repo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
