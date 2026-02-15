package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
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

	// 2. Query Database (Lightweight but includes rooms for ID mapping)
	// We SELECT * FROM hotels WHERE city_id = ?
	// We add a Limit(50) to prevent fetching 10,000 hotels at once
	result := store.DB.Preload("Rooms").Where("city_id = ?", cityID).Limit(50).Find(&hotels)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch hotels",
		})
	}

	// 3. Return Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"count":  len(hotels),
		"data":   hotels,
	})
}
