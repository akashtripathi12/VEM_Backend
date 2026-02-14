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

	log.Println("🗑️  Deleting all hotels from database...")

	// Delete all hotels
	result := store.DB.Exec("DELETE FROM hotels")
	if result.Error != nil {
		log.Fatalf("❌ Error deleting hotels: %v", result.Error)
	}

	log.Printf("✅ Deleted %d hotels", result.RowsAffected)
	log.Println("🎉 Hotels table cleared successfully!")
}
