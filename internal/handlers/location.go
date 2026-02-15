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

// GetCountries fetches all available countries
// GET /api/v1/locations/countries
func (r *Repository) GetCountries(c *fiber.Ctx) error {
	var countries []models.Country
	cacheKey := "locations:countries"
	ctx := context.Background()

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &countries); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return utils.SuccessResponse(c, fiber.StatusOK, countries)
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	} else {
		log.Println("⚠️ [REDIS] SKIPPED: Client not initialized")
	}

	// 2. If not in Redis (or Redis disabled), fetch from DB
	result := store.DB.Order("name ASC").Find(&countries)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch countries")
	}

	// 3. Store in Redis for future requests (Expires in 15 days)
	if store.RDB != nil {
		if data, err := json.Marshal(countries); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s (TTL: 15 days)\n", cacheKey)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, countries)
}

// GetCities fetches cities for a specific country
// GET /api/v1/locations/cities?country_code=AE
func (r *Repository) GetCities(c *fiber.Ctx) error {
	countryCode := c.Query("country_code")

	if countryCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "country_code query parameter is required",
		})
	}

	var cities []models.City
	cacheKey := "locations:cities:" + countryCode
	ctx := context.Background()

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &cities); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return utils.SuccessResponse(c, fiber.StatusOK, cities)
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	// 2. If not in Redis, fetch from DB
	result := store.DB.Where("country_code = ?", countryCode).
		Order("is_popular DESC, name ASC").
		Find(&cities)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch cities")
	}

	// 3. Store in Redis for future requests (Expires in 15 days)
	if store.RDB != nil {
		if data, err := json.Marshal(cities); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s (TTL: 15 days)\n", cacheKey)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, cities)
}
