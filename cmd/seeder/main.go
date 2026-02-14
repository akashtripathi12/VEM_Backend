package main

import (
	"log"
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	store.InitDB()

	log.Println("⚠️  STARTING DATABASE RESET...")

	store.DB.Exec("TRUNCATE TABLE users CASCADE")
	store.DB.Exec("TRUNCATE TABLE guests CASCADE")
	store.DB.Exec("TRUNCATE TABLE agent_profiles CASCADE")
	store.DB.Exec("TRUNCATE TABLE events CASCADE")
	
	log.Println("✅ Database Cleared.")
	// Sync schema to DB
	err := store.DB.AutoMigrate(
        // 1. Auth System
        &models.User{},
        &models.AgentProfile{},

        // 2. Global Location Hierarchy
        &models.Country{},
        &models.City{},

        // 3. Hotel Inventory (The Product)
        &models.Hotel{},
        &models.RoomOffer{},
        &models.BanquetHall{},
        &models.CateringMenu{},

        // 4. Event Management
        &models.Event{},
        &models.Guest{},

        // 5. Allocation Logic (The Join Table)
        &models.GuestAllocation{},
    )

    if err != nil {
        log.Fatal("❌ Migration Failed:", err)
    }
    
    log.Println("✅ All tables created successfully!")

	log.Println("🌱 Seeding new data...")

	// populating logic here

	log.Println("🎉 Database Reset & Populated Successfully!")
}
