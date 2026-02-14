package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

const (
	HotelCodeListURL = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/TBOHotelCodeList"
)

// HotelCodeListRequest represents the API request structure
type HotelCodeListRequest struct {
	CityCode           string `json:"CityCode"`
	IsDetailedResponse string `json:"IsDetailedResponse"`
}

// HotelCodeListResponse represents the API response structure
type HotelCodeListResponse struct {
	Status struct {
		Code        int    `json:"Code"`
		Description string `json:"Description"`
	} `json:"Status"`
	Hotels []struct {
		HotelCode   string `json:"HotelCode"`
		HotelName   string `json:"HotelName"`
		HotelRating string `json:"HotelRating"`
		ImageUrls   []struct {
			ImageUrl string `json:"ImageUrl"`
		} `json:"ImageUrls"`
		Address         string   `json:"Address"`
		HotelLocation   string   `json:"HotelLocation"`
		CountryName     string   `json:"CountryName"`
		CountryCode     string   `json:"CountryCode"`
		Description     string   `json:"Description"`
		HotelFacilities []string `json:"HotelFacilities"`
		Map             string   `json:"Map"`
		Email           string   `json:"Email"`
		PhoneNumber     string   `json:"PhoneNumber"`
		PinCode         string   `json:"PinCode"`
		HotelWebsiteUrl string   `json:"HotelWebsiteUrl"`
		CityName        string   `json:"CityName"`
	} `json:"Hotels"`
}

// SeedHotels fetches hotels for all cities from TBO API and populates the database
func SeedHotels(limitPerCity int) {
	log.Println("🏨 Fetching hotels from TBO API...")

	// Get all cities from database
	var cities []models.City
	result := store.DB.Select("id, name").Find(&cities)
	if result.Error != nil {
		log.Printf("❌ Error fetching cities: %v", result.Error)
		return
	}

	log.Printf("📋 Found %d cities to process", len(cities))

	// Collect ALL hotels from ALL cities first
	var allHotels []models.Hotel
	successfulCities := 0
	skippedCities := 0

	// Fetch hotels for each city
	for i, city := range cities {
		log.Printf("🏙️  [%d/%d] Fetching hotels for: %s (ID: %s)", i+1, len(cities), city.Name, city.ID)

		hotels, err := fetchHotelsForCity(city.ID)
		if err != nil {
			log.Printf("⚠️  Error fetching hotels for %s: %v", city.Name, err)
			continue
		}

		if len(hotels) == 0 {
			skippedCities++
			continue
		}

		// Cap hotels per city
		if len(hotels) > limitPerCity {
			hotels = hotels[:limitPerCity]
		}

		// Add hotels to the collection
		for _, hotel := range hotels {
			// Convert star rating to integer
			starRating := convertStarRating(hotel.HotelRating)

			// Extract image URLs
			var imageUrls []string
			for _, img := range hotel.ImageUrls {
				imageUrls = append(imageUrls, img.ImageUrl)
			}

			// Convert to JSON
			facilitiesJSON, _ := json.Marshal(hotel.HotelFacilities)
			imageUrlsJSON, _ := json.Marshal(imageUrls)

			allHotels = append(allHotels, models.Hotel{
				ID:         hotel.HotelCode,
				CityID:     city.ID,
				Name:       hotel.HotelName,
				StarRating: starRating,
				Address:    hotel.Address,
				Facilities: datatypes.JSON(facilitiesJSON),
				ImageUrls:  datatypes.JSON(imageUrlsJSON),
			})
		}

		successfulCities++
		log.Printf("✅ Collected %d hotels for %s (Total so far: %d)", len(hotels), city.Name, len(allHotels))

		// Small delay to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("📦 Collected %d total hotels from %d cities (%d cities skipped)", len(allHotels), successfulCities, skippedCities)
	log.Println("💾 Starting batch insert into database...")

	// Insert ALL hotels in one massive batch operation
	if len(allHotels) > 0 {
		// Use ON CONFLICT DO NOTHING to skip duplicates instead of failing
		result := store.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(allHotels, 500)
		if result.Error != nil {
			log.Printf("❌ Error batch inserting hotels: %v", result.Error)
			return
		}
		log.Printf("🎉 Successfully inserted %d hotels into database!", result.RowsAffected)
	} else {
		log.Println("⚠️  No hotels to insert")
	}
}

// fetchHotelsForCity fetches hotels for a specific city from TBO API
func fetchHotelsForCity(cityCode string) ([]struct {
	HotelCode   string `json:"HotelCode"`
	HotelName   string `json:"HotelName"`
	HotelRating string `json:"HotelRating"`
	ImageUrls   []struct {
		ImageUrl string `json:"ImageUrl"`
	} `json:"ImageUrls"`
	Address         string   `json:"Address"`
	HotelLocation   string   `json:"HotelLocation"`
	CountryName     string   `json:"CountryName"`
	CountryCode     string   `json:"CountryCode"`
	Description     string   `json:"Description"`
	HotelFacilities []string `json:"HotelFacilities"`
	Map             string   `json:"Map"`
	Email           string   `json:"Email"`
	PhoneNumber     string   `json:"PhoneNumber"`
	PinCode         string   `json:"PinCode"`
	HotelWebsiteUrl string   `json:"HotelWebsiteUrl"`
	CityName        string   `json:"CityName"`
}, error) {
	// Create request body
	reqBody := HotelCodeListRequest{
		CityCode:           cityCode,
		IsDetailedResponse: "true",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", HotelCodeListURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Set Basic Auth
	req.SetBasicAuth(APIUsername, APIPassword)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var apiResponse HotelCodeListResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	// Check API status
	if apiResponse.Status.Code != 200 {
		return nil, nil // Return empty list, not an error
	}

	return apiResponse.Hotels, nil
}

// convertStarRating converts string rating to integer
func convertStarRating(rating string) int {
	switch rating {
	case "OneStar":
		return 1
	case "TwoStar":
		return 2
	case "ThreeStar":
		return 3
	case "FourStar":
		return 4
	case "FiveStar":
		return 5
	default:
		return 0
	}
}
