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

// GetHotelsByCity fetches the raw hotel list for a city (No rooms, just hotel info)
// GET /api/v1/hotels?city_id=DXB
func (r *Repository) GetHotelsByCity(c *fiber.Ctx) error {
	cityID := c.Query("city_id")

	// 1. Validate Input
	if cityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "city_id query parameter is required",
		})
	}

	var hotels []models.Hotel
	cacheKey := "hotels:city:" + cityID
	ctx := context.Background()

	// 2. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &hotels); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"status": "success",
					"count":  len(hotels),
					"data":   hotels,
				})
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	// 3. Query Database
	result := store.DB.Preload("Rooms").Where("city_id = ?", cityID).Limit(50).Find(&hotels)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch hotels")
	}

	// 4. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(hotels); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	// 5. Return Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"count":  len(hotels),
		"data":   hotels,
	})
}
