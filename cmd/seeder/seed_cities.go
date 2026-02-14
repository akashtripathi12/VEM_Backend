package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
)

const (
	CityListURL = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/CityList"
)

// CityListRequest represents the API request structure
type CityListRequest struct {
	CountryCode string `json:"CountryCode"`
}

// CityListResponse represents the API response structure
type CityListResponse struct {
	Status struct {
		Code        int    `json:"Code"`
		Description string `json:"Description"`
	} `json:"Status"`
	CityList []struct {
		Code      string `json:"Code"`
		Name      string `json:"Name"`
		IsPopular bool   `json:"IsPopular"`
	} `json:"CityList"`
}

// SeedCities fetches cities for all countries from TBO API and populates the database
func SeedCities() {
	log.Println("🏙️  Fetching cities from TBO API...")

	// Get all country codes from database
	var countries []models.Country
	result := store.DB.Select("code").Find(&countries)
	if result.Error != nil {
		log.Printf("❌ Error fetching countries: %v", result.Error)
		return
	}

	log.Printf("📋 Found %d countries to process", len(countries))

	// Collect ALL cities from ALL countries first
	var allCities []models.City
	successfulCountries := 0

	// Fetch cities for each country
	for i, country := range countries {
		log.Printf("🌍 [%d/%d] Fetching cities for: %s", i+1, len(countries), country.Code)

		cities, err := fetchCitiesForCountry(country.Code)
		if err != nil {
			log.Printf("⚠️  Error fetching cities for %s: %v", country.Code, err)
			continue
		}

		if len(cities) == 0 {
			continue
		}

		// Add cities to the collection
		for _, city := range cities {
			allCities = append(allCities, models.City{
				ID:          city.Code,
				CountryCode: country.Code,
				Name:        city.Name,
				IsPopular:   city.IsPopular,
			})
		}

		successfulCountries++
		log.Printf("✅ Collected %d cities for %s (Total so far: %d)", len(cities), country.Code, len(allCities))

		// Small delay to avoid overwhelming the API
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("📦 Collected %d total cities from %d countries", len(allCities), successfulCountries)
	log.Println("💾 Starting batch insert into database...")

	// Insert ALL cities in one massive batch operation
	if len(allCities) > 0 {
		result := store.DB.CreateInBatches(allCities, 500)
		if result.Error != nil {
			log.Printf("❌ Error batch inserting cities: %v", result.Error)
			return
		}
		log.Printf("🎉 Successfully inserted %d cities into database!", result.RowsAffected)
	} else {
		log.Println("⚠️  No cities to insert")
	}
}

// fetchCitiesForCountry fetches cities for a specific country from TBO API
func fetchCitiesForCountry(countryCode string) ([]struct {
	Code      string `json:"Code"`
	Name      string `json:"Name"`
	IsPopular bool   `json:"IsPopular"`
}, error) {
	// Create request body
	reqBody := CityListRequest{
		CountryCode: countryCode,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", CityListURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Set Basic Auth
	req.SetBasicAuth(APIUsername, APIPassword)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read raw response for debugging
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, err
	}

	// Debug: Print first response to understand structure
	if countryCode == "IN" || countryCode == "US" || countryCode == "AE" {
		jsonBytes, _ := json.MarshalIndent(rawResponse, "", "  ")
		log.Printf("🔍 DEBUG - API Response for %s:\n%s", countryCode, string(jsonBytes))
	}

	// Parse response
	var apiResponse CityListResponse
	jsonBytes, _ := json.Marshal(rawResponse)
	if err := json.Unmarshal(jsonBytes, &apiResponse); err != nil {
		return nil, err
	}

	// Check API status
	if apiResponse.Status.Code != 200 {
		// Don't log for every country, too verbose
		return nil, nil // Return empty list, not an error
	}

	return apiResponse.CityList, nil
}
