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

	log.Println("🎲 Updating hotel occupancy with random values (200-1000)...")

	// Update all hotels in a single query using SQL's random function
	// FLOOR(RANDOM() * 801) + 200 generates a random integer between 200 and 1000
	result := store.DB.Exec("UPDATE hotels SET occupancy = FLOOR(RANDOM() * 801) + 200")
	if result.Error != nil {
		log.Fatalf("❌ Error updating hotel occupancy: %v", result.Error)
	}

	log.Printf("🎉 Successfully updated %d hotels with random occupancy values!", result.RowsAffected)
}
