package main

import (
	"log"
	"strings" // Added for joining allowed origins

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/handlers"
	"github.com/akashtripathi12/TBO_Backend/internal/queue"
	"github.com/akashtripathi12/TBO_Backend/internal/routes"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables from .env file (if it exists)
	godotenv.Load()

	// 2. Load Centralized Config (All os.Getenv happens INSIDE here)
	cfg := config.Load()

	// Initialize Store
	store.InitDB()
	log.Println("✅ DB Connected")

	// Initialize Redis
	store.InitRedis(cfg)

	// --- Asynq Redis Config (Using cfg!) ---
	// Make sure your config.go maps REDIS_URL to cfg.RedisAddr
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("❌ Invalid Redis URL: %v", err)
	}

	// Initialize Asynq Client (Producer)
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// Initialize Repository with Queue Client
	repo := handlers.NewRepository(cfg, store.DB, client)

	// Initialize Asynq Server (Consumer)
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	// Register Task Handlers
	handler := &queue.TaskHandler{Cfg: cfg}
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeEmailDelivery, handler.HandleEmailTask)

	// Run Worker in Background
	go func() {
		log.Println("👷 Asynq Worker Server Starting...")
		if err := srv.Run(mux); err != nil {
			log.Printf("❌ Asynq Server Failed: %v", err)
		}
	}()

	app := fiber.New()

	// Enable CORS (Using cfg!)
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(cfg.AllowedOrigins, ", "),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup Routes
	routes.SetupRoutes(app, cfg, repo)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("TBO Backend Operational 🚀")
	})

	// Start Server (Using cfg! Note: config.go already added the ":" to Port)
	log.Fatal(app.Listen(cfg.Port))
}
