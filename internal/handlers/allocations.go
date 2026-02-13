package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (m *Repository) CreateAllocation(c *fiber.Ctx) error {
	// TODO: Parse request body
	// var allocation models.RoomAllocation
	// if err := c.BodyParser(&allocation); err != nil {
	//     return utils.ValidationErrorResponse(c, "Invalid request body")
	// }

	// TODO: Create allocation
	// if err := m.DB.CreateAllocation(allocation); err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to create allocation")
	// }

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Create Allocation Endpoint",
	})
}

func (m *Repository) UpdateAllocation(c *fiber.Ctx) error {
	// Get allocation ID from path parameter
	id := c.Params("id")

	// TODO: Parse request body and update allocation
	// var allocation models.RoomAllocation
	// if err := c.BodyParser(&allocation); err != nil {
	//     return utils.ValidationErrorResponse(c, "Invalid request body")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Update Allocation Endpoint",
		"id":      id,
	})
}
