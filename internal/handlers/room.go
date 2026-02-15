package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetRoomsByHotel fetches all available room offers for a specific hotel
// GET /api/v1/hotels/:hotelCode/rooms
func (r *Repository) GetRoomsByHotel(c *fiber.Ctx) error {
	hotelCode := c.Params("hotelCode")

	// 1. Validate Input
	if hotelCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "hotelCode parameter is required",
		})
	}

	var rooms []models.RoomOffer
	cacheKey := "rooms:hotel:" + hotelCode
	ctx := context.Background()

	// 2. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &rooms); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"status": "success",
					"count":  len(rooms),
					"data":   rooms,
				})
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	// 3. Query Database
	result := store.DB.Where("hotel_id = ?", hotelCode).Find(&rooms)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch rooms")
	}

	// 4. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(rooms); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	// 5. Return Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"count":  len(rooms),
		"data":   rooms,
	})
}
