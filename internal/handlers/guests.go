package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// List Event Guests
func (m *Repository) GetGuests(c *fiber.Ctx) error {
	eventID := c.Params("id")
	var guests []models.Guest

	if err := m.DB.Where("event_id = ?", eventID).Find(&guests).Error; err != nil {
		return utils.InternalErrorResponse(c, "Failed to fetch guests")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"guests": guests,
	})
}

// Get Single Guest
func (m *Repository) GetGuest(c *fiber.Ctx) error {
	id := c.Params("id")
	var guest models.Guest

	if err := m.DB.First(&guest, "id = ?", id).Error; err != nil {
		return utils.NotFoundResponse(c, "Guest")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"guest": guest,
	})
}



// Create Guest (Generic)
func (m *Repository) CreateGuest(c *fiber.Ctx) error {
	var guest models.Guest
	if err := c.BodyParser(&guest); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := m.DB.Create(&guest).Error; err != nil {
		return utils.InternalErrorResponse(c, "Failed to create guest")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Guest created successfully",
		"guest":   guest,
	})
}

// Add Sub Guest (Not in immediate scope but good to keep generic)
func (m *Repository) AddSubGuest(c *fiber.Ctx) error {
	// Logic to link sub-guest to head guest would go here
	// For now, just create a guest
	return m.CreateGuest(c)
}

// Update Guest
func (m *Repository) UpdateGuest(c *fiber.Ctx) error {
	id := c.Params("id")
	var input models.Guest

	if err := c.BodyParser(&input); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	var guest models.Guest
	if err := m.DB.First(&guest, "id = ?", id).Error; err != nil {
		return utils.NotFoundResponse(c, "Guest")
	}

	// Update fields
	m.DB.Model(&guest).Updates(input)

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Guest updated successfully",
		"guest":   guest,
	})
}

// Delete Guest
func (m *Repository) DeleteGuest(c *fiber.Ctx) error {
	id := c.Params("id")

	// Transaction to delete guest and release room (future scope: implementation plan mentioned shadow inventory release)
	// For now, simple delete
	if err := m.DB.Delete(&models.Guest{}, "id = ?", id).Error; err != nil {
		return utils.InternalErrorResponse(c, "Failed to delete guest")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Guest deleted successfully",
	})
}
