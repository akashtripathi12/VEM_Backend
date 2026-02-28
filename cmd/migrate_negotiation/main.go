package main

import (
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/joho/godotenv"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	store.InitDB()
	db := store.DB
	db.Logger = logger.Default.LogMode(logger.Info)

	log.Println("🔄 Migrating negotiation tables...")

	err := db.AutoMigrate(
		&models.NegotiationSession{},
		&models.NegotiationRound{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Negotiation tables created successfully!")
}
