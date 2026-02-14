package store

import (
	"log"
	"os"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to Supabase")

	// Auto Migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.AgentProfile{},
		&models.Guest{},
		&models.Event{},
		&models.Country{},
		&models.City{},
		&models.Hotel{},
		&models.RoomOffer{},
		&models.BanquetHall{},
		&models.CateringMenu{},
		&models.GuestAllocation{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database Migrated")
}
