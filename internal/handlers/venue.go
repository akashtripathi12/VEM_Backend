package handlers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetBanquetsByHotel retrieves banquet halls for a specific hotel, optionally filtered by capacity
func (r *Repository) GetBanquetsByHotel(c *fiber.Ctx) error {
	hotelCode := c.Params("hotelCode")
	capacityStr := c.Query("capacity")

	var banquets []models.BanquetHall
	cacheKey := "banquets:hotel:" + hotelCode + ":cap:" + capacityStr
	ctx := context.Background()

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &banquets); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return c.JSON(banquets)
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	query := store.DB.Where("hotel_id = ?", hotelCode)

	if capacityStr != "" {
		capacity, err := strconv.Atoi(capacityStr)
		if err == nil {
			query = query.Where("capacity >= ?", capacity)
		}
	}

	if err := query.Find(&banquets).Error; err != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch banquets")
	}

	// 2. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(banquets); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	return c.JSON(banquets)
}

// GetCateringByHotel retrieves catering menus for a specific hotel
func (r *Repository) GetCateringByHotel(c *fiber.Ctx) error {
	hotelCode := c.Params("hotelCode")
	var menus []models.CateringMenu
	cacheKey := "catering:hotel:" + hotelCode
	ctx := context.Background()

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &menus); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return c.JSON(menus)
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	if err := store.DB.Where("hotel_id = ?", hotelCode).Find(&menus).Error; err != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch catering menus")
	}

	// 2. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(menus); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 15*24*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	return c.JSON(menus)
}
