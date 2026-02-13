package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (m *Repository) GetGuests(c *fiber.Ctx) error {
	// Get event ID from path parameter
	id := c.Params("id")

	// TODO: Get guests by event ID
	// guests, err := m.DB.GetGuestsByEventID(id)
	// if err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to fetch guests")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Get Guests Endpoint",
		"eventId": id,
		"guests":  []interface{}{},
	})
}

func (m *Repository) GetGuest(c *fiber.Ctx) error {
	// Get guest ID from path parameter
	id := c.Params("id")

	// TODO: Get guest by ID
	// guest, err := m.DB.GetGuestByID(id)
	// if err != nil {
	//     return utils.NotFoundResponse(c, "Guest")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Get Guest Endpoint",
		"id":      id,
	})
}

func (m *Repository) CreateGuest(c *fiber.Ctx) error {
	// TODO: Parse request body
	// var guest models.HeadGuest
	// if err := c.BodyParser(&guest); err != nil {
	//     return utils.ValidationErrorResponse(c, "Invalid request body")
	// }

	// TODO: Create guest
	// if err := m.DB.AddHeadGuest(guest); err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to create guest")
	// }

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Create Guest Endpoint",
	})
}

func (m *Repository) AddSubGuest(c *fiber.Ctx) error {
	// Get head guest ID from path parameter
	id := c.Params("id")

	// TODO: Parse request body
	// var subGuest models.SubGuest
	// if err := c.BodyParser(&subGuest); err != nil {
	//     return utils.ValidationErrorResponse(c, "Invalid request body")
	// }

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message":     "Add Sub Guest Endpoint",
		"headGuestId": id,
	})
}

func (m *Repository) UpdateGuest(c *fiber.Ctx) error {
	// Get guest ID from path parameter
	id := c.Params("id")

	// TODO: Parse request body and update guest
	// var guest models.HeadGuest
	// if err := c.BodyParser(&guest); err != nil {
	//     return utils.ValidationErrorResponse(c, "Invalid request body")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Update Guest Endpoint",
		"id":      id,
	})
}

func (m *Repository) DeleteGuest(c *fiber.Ctx) error {
	// Get guest ID from path parameter
	id := c.Params("id")

	// TODO: Delete guest
	// if err := m.DB.DeleteGuest(id); err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to delete guest")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Delete Guest Endpoint",
		"id":      id,
	})
}
