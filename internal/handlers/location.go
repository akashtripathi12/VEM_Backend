package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetCountries fetches all available countries
// GET /api/v1/locations/countries
func (r *Repository) GetCountries(c *fiber.Ctx) error {
	var countries []models.Country

	// Fetch all countries, sorted by name
	result := store.DB.Order("name ASC").Find(&countries)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch countries")
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

	// Fetch cities for this country, popular ones first
	result := store.DB.Where("country_code = ?", countryCode).
		Order("is_popular DESC, name ASC").
		Find(&cities)

	if result.Error != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch cities")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, cities)
}
