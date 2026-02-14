package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

const (
	RoomSearchURL = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/search"
)

// RoomSearchRequest represents the API request structure
type RoomSearchRequest struct {
	CheckIn            string    `json:"CheckIn"`
	CheckOut           string    `json:"CheckOut"`
	HotelCodes         string    `json:"HotelCodes"`
	GuestNationality   string    `json:"GuestNationality"`
	PaxRooms           []PaxRoom `json:"PaxRooms"`
	ResponseTime       int       `json:"ResponseTime"`
	IsDetailedResponse bool      `json:"IsDetailedResponse"`
}

type PaxRoom struct {
	Adults int `json:"Adults"`
}

// RoomSearchResponse represents the API response structure
type RoomSearchResponse struct {
	Status struct {
		Code        int    `json:"Code"`
		Description string `json:"Description"`
	} `json:"Status"`
	HotelResult []struct {
		HotelCode string `json:"HotelCode"`
		Currency  string `json:"Currency"`
		Rooms     []struct {
			Name           []string `json:"Name"`
			BookingCode    string   `json:"BookingCode"`
			Inclusion      string   `json:"Inclusion"`
			TotalFare      float64  `json:"TotalFare"`
			TotalTax       float64  `json:"TotalTax"`
			CancelPolicies []struct {
				FromDate           string  `json:"FromDate"`
				ChargeType         string  `json:"ChargeType"`
				CancellationCharge float64 `json:"CancellationCharge"`
			} `json:"CancelPolicies"`
			IsRefundable bool `json:"IsRefundable"`
		} `json:"Rooms"`
	} `json:"HotelResult"`
}

// SeedRooms fetches rooms for hotels and populates the database
func SeedRooms(limit int) {
	log.Printf("🛏️  Fetching rooms from TBO API (Limit: %d hotels)...", limit)

	// Step 0: Clear existing data and Update Schema
	log.Println("🧹 Clearing existing room offers and updating schema...")

	// Drop the table to ensure schema changes (ID uint -> string) are applied correctly.
	// We use CASCADE to drop dependent foreign keys (like in guest_allocations) if they exist.
	if err := store.DB.Exec("DROP TABLE IF EXISTS room_offers CASCADE").Error; err != nil {
		log.Printf("⚠️  Error dropping table: %v", err)
	}

	// Re-create table with new schema
	if err := store.DB.AutoMigrate(&models.RoomOffer{}); err != nil {
		log.Printf("❌ Error migrating table room_offers: %v", err)
		return
	}

	// Also migrate GuestAllocation because it references RoomOffer and we changed the ID type
	// This might fail if the table has data with integer IDs, but assuming clean slate or it handles it.
	// Ideally we should drop it too if we want a clean slate.
	store.DB.Exec("DROP TABLE IF EXISTS guest_allocations CASCADE") // Optional but safer for this big refactor
	if err := store.DB.AutoMigrate(&models.GuestAllocation{}); err != nil {
		log.Printf("⚠️ Error migrating guest_allocations: %v", err)
	}

	log.Println("✅ Room offers table recreated and schema updated.")

	// Get hotels from database
	var hotels []models.Hotel
	// ID is mapped to hotel_code in the model
	query := store.DB.Select("hotel_code, name").Order("hotel_code")
	if limit > 0 {
		query = query.Limit(limit)
	}
	result := query.Find(&hotels)
	if result.Error != nil {
		log.Printf("❌ Error fetching hotels: %v", result.Error)
		return
	}

	log.Printf("📋 Found %d hotels to process", len(hotels))

	var batchRoomOffers []models.RoomOffer
	successfulHotels := 0
	totalInserted := int64(0)

	const BatchSaveSize = 20 // Save every 20 hotels (or whenever buffer gets full)

	for i, hotel := range hotels {
		log.Printf("🏨 [%d/%d] Fetching rooms for: %s (ID: %s)", i+1, len(hotels), hotel.Name, hotel.ID)

		currency, rooms, err := fetchRoomsForHotel(hotel.ID)
		if err != nil {
			log.Printf("⚠️  Error fetching rooms for hotel %s: %v", hotel.ID, err)
			continue
		}

		if len(rooms) == 0 {
			// even if no rooms, we continue
			// log.Printf("   No rooms found for %s", hotel.Name)
		} else {
			// Process and collect rooms
			for _, room := range rooms {
				// Convert CancelPolicies to JSON
				policiesJSON, _ := json.Marshal(room.CancelPolicies)

				// Helper to get first name if available
				roomName := "Standard Room"
				if len(room.Name) > 0 {
					roomName = room.Name[0]
				}

				// Calculate total price (Fare + Tax)
				totalPrice := room.TotalFare + room.TotalTax

				batchRoomOffers = append(batchRoomOffers, models.RoomOffer{
					ID:             uuid.New().String(), // Generate unique UUID
					HotelID:        hotel.ID,
					Name:           roomName,
					BookingCode:    room.BookingCode,
					TotalFare:      totalPrice,
					Currency:       currency, // Use currency from API
					IsRefundable:   room.IsRefundable,
					CancelPolicies: datatypes.JSON(policiesJSON),
				})
			}
			successfulHotels++
			log.Printf("✅ Found %d rooms for %s", len(rooms), hotel.Name)
		}

		// Check if batch is ready to save
		if len(batchRoomOffers) >= 100 { // Save every ~100 records
			count := saveRoomBatch(batchRoomOffers)
			totalInserted += count
			batchRoomOffers = nil // Clear slice
			log.Printf("💾 Saved batch of %d rooms (Total inserted: %d)", count, totalInserted)
		}

		// Small delay to avoid overwhelming the API
		time.Sleep(200 * time.Millisecond)
	}

	// Save remaining rooms
	if len(batchRoomOffers) > 0 {
		count := saveRoomBatch(batchRoomOffers)
		totalInserted += count
		log.Printf("💾 Saved final batch of %d rooms", count)
	}

	log.Printf("🎉 Finished! Processed %d hotels. Total rooms inserted: %d", successfulHotels, totalInserted)
}

func saveRoomBatch(rooms []models.RoomOffer) int64 {
	if len(rooms) == 0 {
		return 0
	}
	result := store.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(rooms, 100) // Inner batching for GORM
	if result.Error != nil {
		log.Printf("❌ Error batch inserting rooms: %v", result.Error)
		return 0
	}
	return result.RowsAffected
}

func fetchRoomsForHotel(hotelCode string) (string, []struct {
	Name           []string `json:"Name"`
	BookingCode    string   `json:"BookingCode"`
	Inclusion      string   `json:"Inclusion"`
	TotalFare      float64  `json:"TotalFare"`
	TotalTax       float64  `json:"TotalTax"`
	CancelPolicies []struct {
		FromDate           string  `json:"FromDate"`
		ChargeType         string  `json:"ChargeType"`
		CancellationCharge float64 `json:"CancellationCharge"`
	} `json:"CancelPolicies"`
	IsRefundable bool `json:"IsRefundable"`
}, error) {
	// Create request body
	// Create request body
	reqBody := RoomSearchRequest{
		CheckIn:          "2026-10-01",
		CheckOut:         "2026-10-04",
		HotelCodes:       hotelCode,
		GuestNationality: "AE",
		PaxRooms: []PaxRoom{
			{Adults: 1},
		},
		ResponseTime:       0,
		IsDetailedResponse: true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", RoomSearchURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, err
	}

	// Set Basic Auth
	req.SetBasicAuth(APIUsername, APIPassword)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var apiResponse RoomSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", nil, err
	}

	// Check API status
	if apiResponse.Status.Code != 200 {
		return "", nil, fmt.Errorf("API error: %s (Code: %d)", apiResponse.Status.Description, apiResponse.Status.Code)
	}

	if len(apiResponse.HotelResult) == 0 {
		return "", nil, nil
	}

	// Return Currency and Rooms
	return apiResponse.HotelResult[0].Currency, apiResponse.HotelResult[0].Rooms, nil
}
