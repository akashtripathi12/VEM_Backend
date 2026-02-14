package main

import (
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	store.InitDB()

	// ========== COMMENTED OUT TO PRESERVE EXISTING DATA ==========
	// Uncomment below if you want to reset the database

	// log.Println("⚠️  STARTING DATABASE RESET...")
	// store.DB.Exec("TRUNCATE TABLE users CASCADE")
	// store.DB.Exec("TRUNCATE TABLE guests CASCADE")
	// store.DB.Exec("TRUNCATE TABLE agent_profiles CASCADE")
	// store.DB.Exec("TRUNCATE TABLE events CASCADE")
	// log.Println("✅ Database Cleared.")
	// ========== COMMENTED OUT TO AVOID RE-CREATING TABLES ==========
	// Uncomment below if you need to run migrations

	// err := store.DB.AutoMigrate(
	//     // 1. Auth System
	//     &models.User{},
	//     &models.AgentProfile{},
	//     // 2. Global Location Hierarchy
	//     &models.Country{},
	//     &models.City{},
	//     // 3. Hotel Inventory (The Product)
	//     &models.Hotel{},
	//     &models.RoomOffer{},
	//     &models.BanquetHall{},
	//     &models.CateringMenu{},
	//     // 4. Event Management
	//     &models.Event{},
	//     &models.Guest{},
	//     // 5. Allocation Logic (The Join Table)
	//     &models.GuestAllocation{},
	// )
	// if err != nil {
	//     log.Fatal("❌ Migration Failed:", err)
	// }
	// log.Println("✅ All tables created successfully!")

	log.Println("🌱 Seeding data...")

	// Seed countries from TBO API (comment out if already seeded)
	// SeedCountries()

	// Seed cities from TBO API
	// SeedCities()

	// Seed hotels from TBO API (max 10 per city)
	// SeedHotels()

	// Seed rooms for ALL hotels
	// SeedRooms(0)

	log.Println("🎉 Data Seeding Completed Successfully!")
}
