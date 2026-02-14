package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
)

const (
	CountryListURL = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/CountryList"
	APIUsername    = "hackathontest"
	APIPassword    = "Hac@98147521"
)

// CountryResponse represents the API response structure
type CountryResponse struct {
	Status struct {
		Code        int    `json:"Code"`
		Description string `json:"Description"`
	} `json:"Status"`
	Countries []struct {
		Code      string `json:"Code"`
		Name      string `json:"Name"`
		PhoneCode string `json:"PhoneCode"`
	} `json:"CountryList"`
}

// SeedCountries fetches countries from TBO API and populates the database
func SeedCountries() {
	log.Println("🌍 Fetching countries from TBO API...")

	// Create HTTP request
	req, err := http.NewRequest("GET", CountryListURL, nil)
	if err != nil {
		log.Printf("❌ Error creating request: %v", err)
		return
	}

	// Set Basic Auth
	req.SetBasicAuth(APIUsername, APIPassword)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	// Parse response
	var apiResponse CountryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Printf("❌ Error decoding response: %v", err)
		return
	}

	// Check API status
	if apiResponse.Status.Code != 200 {
		log.Printf("❌ API Error: %s (Code: %d)", apiResponse.Status.Description, apiResponse.Status.Code)
		return
	}

	log.Printf("✅ Fetched %d countries from API", len(apiResponse.Countries))

	// Prepare all countries for batch insert
	var countries []models.Country
	for _, country := range apiResponse.Countries {
		countries = append(countries, models.Country{
			Code:      country.Code,
			Name:      country.Name,
			PhoneCode: country.PhoneCode,
		})
	}

	// Batch insert all countries at once
	// Using CreateInBatches for better performance with large datasets
	result := store.DB.CreateInBatches(countries, 100)
	if result.Error != nil {
		log.Printf("❌ Error batch inserting countries: %v", result.Error)
		return
	}

	log.Printf("✅ Successfully seeded %d countries into database", result.RowsAffected)
}
