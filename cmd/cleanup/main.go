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

	log.Println("🗑️  Cleaning up database to keep only 10 target countries...")

	// Target countries: US, DE, SG, AE, GB, ES, TH, CN, AU, IN
	targetCountries := []string{"US", "SG", "AE", "TH", "IN"}

	// Step 1: Delete cities NOT from target countries
	log.Println("🏙️  Deleting cities from non-target countries...")
	result := store.DB.Exec("DELETE FROM cities WHERE country_code NOT IN ?", targetCountries)
	if result.Error != nil {
		log.Fatalf("❌ Error deleting cities: %v", result.Error)
	}
	log.Printf("✅ Deleted %d cities from non-target countries", result.RowsAffected)

	// Step 2: Delete countries NOT in target list
	log.Println("🌍 Deleting non-target countries...")
	result = store.DB.Exec("DELETE FROM countries WHERE code NOT IN ?", targetCountries)
	if result.Error != nil {
		log.Fatalf("❌ Error deleting countries: %v", result.Error)
	}
	log.Printf("✅ Deleted %d countries", result.RowsAffected)

	// Step 3: Cap cities to 100 per country
	log.Println("\n🔢 Capping cities to 100 per country...")

	for _, countryCode := range targetCountries {
		// Delete cities beyond the 1000 limit for this country
		result := store.DB.Exec(`
			DELETE FROM cities 
			WHERE country_code = ? 
			AND id NOT IN (
				SELECT id FROM cities 
				WHERE country_code = ? 
				ORDER BY id 
				LIMIT 100
			)
		`, countryCode, countryCode)

		if result.Error != nil {
			log.Printf("⚠️  Error capping cities for %s: %v", countryCode, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("   %s: Removed %d cities (kept 100)", countryCode, result.RowsAffected)
		}
	}

	// Verification
	log.Println("\n📊 Final Verification:")

	var countryCount int64
	store.DB.Model(&struct {
		Code string `gorm:"primaryKey"`
	}{}).Table("countries").Count(&countryCount)
	log.Printf("✅ Total countries: %d", countryCount)

	var cityCount int64
	store.DB.Model(&struct {
		ID string `gorm:"primaryKey"`
	}{}).Table("cities").Count(&cityCount)
	log.Printf("✅ Total cities: %d", cityCount)

	// Show cities per country
	log.Println("\n📍 Cities per country (max 100 each):")
	var results []struct {
		CountryCode string
		Count       int64
	}
	store.DB.Table("cities").
		Select("country_code, COUNT(*) as count").
		Group("country_code").
		Order("country_code").
		Scan(&results)

	for _, r := range results {
		log.Printf("   %s: %d cities", r.CountryCode, r.Count)
	}

	log.Println("\n🎉 Database cleanup completed successfully!")
}
