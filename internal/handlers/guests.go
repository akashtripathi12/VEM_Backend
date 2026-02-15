package handlers

import (
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	eventID := c.Params("id")

	// Explicit Input Struct to interpret JSON strictly
	type GuestInput struct {
		Name          string       `json:"name"`
		Age           int          `json:"age"`
		Type          string       `json:"type"`
		Phone         string       `json:"phone"`
		Email         string       `json:"email"`
		ArrivalDate   time.Time    `json:"arrivalDate"`
		DepartureDate time.Time    `json:"departureDate"`
		FamilyMembers []GuestInput `json:"family_members"`
	}

	// DEBUG: Print raw body
	rawBody := c.Body()
	// fmt.Println("DEBUG RAW BODY:", string(rawBody))

	// 1. Try generic map parsing to check JSON syntax
	var genericMap map[string]interface{}
	if err := c.BodyParser(&genericMap); err != nil {
		return utils.ValidationErrorResponse(c, "DEBUG: JSON Syntax Error: "+err.Error()+ " | Body: "+string(rawBody))
	}

	var req GuestInput
	if err := c.BodyParser(&req); err != nil {
		// Include err.Error() to help debug validity issues
		return utils.ValidationErrorResponse(c, "DEBUG: Struct Mapping Error: "+err.Error())
	}

	// Basic Validation
	if req.Name == "" {
		return utils.ValidationErrorResponse(c, "Name is required")
	}

	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid Event ID")
	}

	// 🔥 Generate ONE FamilyID for this registration
	familyID := uuid.New()

	// Helper to convert Input -> Model
	toModel := func(input GuestInput) models.Guest {
		return models.Guest{
			ID:            uuid.New(),
			EventID:       parsedEventID,
			FamilyID:      familyID,
			Name:          input.Name,
			Age:           input.Age,
			Type:          input.Type, // Will be auto-filled if empty below
			Phone:         input.Phone,
			Email:         input.Email,
			ArrivalDate:   input.ArrivalDate,
			DepartureDate: input.DepartureDate,
		}
	}

	// Prepare list of all guests to save
	var allGuests []models.Guest

	// 1. Add Main Guest
	mainGuest := toModel(req)
	allGuests = append(allGuests, mainGuest)

	// 2. Add Family Members
	if len(req.FamilyMembers) > 0 {
		for _, memberInput := range req.FamilyMembers {
			memberModel := toModel(memberInput)
			
			// Inherit dates from main guest if missing
			if memberModel.ArrivalDate.IsZero() {
				memberModel.ArrivalDate = mainGuest.ArrivalDate
			}
			if memberModel.DepartureDate.IsZero() {
				memberModel.DepartureDate = mainGuest.DepartureDate
			}

			allGuests = append(allGuests, memberModel)
		}
	}

	tx := m.DB.Begin()

	for i := range allGuests {
		// Auto detect type is handled by BeforeSave hook if Type is empty

		if err := tx.Create(&allGuests[i]).Error; err != nil {
			tx.Rollback()
			return utils.InternalErrorResponse(c, "Failed to create guests")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.InternalErrorResponse(c, "Transaction failed")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message":     "Registration successful",
		"family_id":   familyID,
		"total_guest": len(allGuests),
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
// Delete Guest
func (m *Repository) DeleteGuest(c *fiber.Ctx) error {
	id := c.Params("id")

	// Transaction to delete guest and release room (future scope: implementation plan mentioned shadow inventory release)
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Delete associated GuestAllocations first to avoid FK violation
	if err := tx.Where("guest_id = ?", id).Delete(&models.GuestAllocation{}).Error; err != nil {
		tx.Rollback()
		return utils.InternalErrorResponse(c, "Failed to delete guest allocations")
	}

	// 2. Delete the guest
	if err := tx.Delete(&models.Guest{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return utils.InternalErrorResponse(c, "Failed to delete guest")
	}

	// 3. Commit
	if err := tx.Commit().Error; err != nil {
		return utils.InternalErrorResponse(c, "Transaction failed")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Guest deleted successfully",
	})
}
