package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/queue"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (m *Repository) GetEvents(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	// Define response struct locally for the handler
	type EventResponse struct {
		models.Event
		GuestCount           int64    `json:"guestCount"`
		InventoryConsumed    float64  `json:"inventoryConsumed"`
		BudgetSpent          float64  `json:"budgetSpent"`
		TotalBudget          *float64 `json:"totalBudget"`
		DaysUntilEvent       int      `json:"daysUntilEvent"`
		PendingActions       int      `json:"pendingActions"`
		PendingActionDetails []string `json:"pendingActionDetails"`
	}

	// userID is stored as uuid.UUID in context by middleware
	agentID, ok := userID.(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid User ID type")
	}

	var events []models.Event
	cacheKey := fmt.Sprintf("events:agent:%s", agentID.String())
	ctx := context.Background()

	var responseEvents []EventResponse

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &responseEvents); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
					"message": "Events Fetched Successfully (Cached)",
					"events":  responseEvents,
				})
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	if err := m.DB.Where("agent_id = ?", agentID).Find(&events).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch events")
	}

	// Fetch Agent Name
	var agent models.User
	agentName := "Agent"
	if err := m.DB.Where("id = ?", agentID).First(&agent).Error; err == nil {
		agentName = agent.Name
	}

	// 3. Decorate with Metrics
	for _, evt := range events {
		evt.Organizer = agentName
		var guestCount int64
		m.DB.Model(&models.Guest{}).Where("event_id = ?", evt.ID).Count(&guestCount)

		var allocatedCount int64
		m.DB.Model(&models.GuestAllocation{}).Where("event_id = ?", evt.ID).Count(&allocatedCount)

		// Calculate basic inventory consumption %
		// (Total allocated / Total Available)
		var consumption float64 = 0.0
		if len(evt.RoomsInventory) > 0 {
			var totalRooms int = 0
			// Unmarshal if JSON was read into a raw string, but it's datatypes.JSON
			var rooms []RoomsInventoryItem
			json.Unmarshal(evt.RoomsInventory, &rooms)

			for _, rp := range rooms {
				totalRooms += rp.Available // Or rp.Total
			}
			if totalRooms > 0 {
				consumption = (float64(allocatedCount) / float64(totalRooms)) * 100
			}
		}

		// Calculate Days Until Event
		daysUntil := int(time.Until(evt.StartDate).Hours() / 24)

		// Calculate Pending Actions (e.g. guests without allocations)
		var pendingGuests int64
		if evt.Status != "locked" && evt.Status != "finalized" {
			m.DB.Model(&models.Guest{}).
				Joins("LEFT JOIN guest_allocations ON guests.id = guest_allocations.guest_id").
				Where("guests.event_id = ? AND guest_allocations.id IS NULL", evt.ID).
				Count(&pendingGuests)
		}

		// And maybe lacking a head guest
		pendingActions := int(pendingGuests)
		var pendingActionDetails []string
		if pendingGuests > 0 {
			pendingActionDetails = append(pendingActionDetails, fmt.Sprintf("%d guest(s) need room allocation", pendingGuests))
		}
		if evt.HeadGuestID == uuid.Nil {
			pendingActions++
			pendingActionDetails = append(pendingActionDetails, "Assign a Head Guest")
		}

		// Budget Spent: prefer stored value (set by Make Payment), else compute from allocations
		var budgetSpent float64
		if evt.BudgetSpent != nil && *evt.BudgetSpent > 0 {
			budgetSpent = *evt.BudgetSpent
		} else {
			m.DB.Model(&models.GuestAllocation{}).Where("event_id = ?", evt.ID).Select("COALESCE(SUM(locked_price), 0)").Scan(&budgetSpent)
		}

		// If no budget is set but we have budgetSpent, use budgetSpent as totalBudget so bar fills to 100%
		totalBudget := evt.Budget
		if totalBudget == nil && budgetSpent > 0 {
			totalBudget = &budgetSpent
		}

		responseEvents = append(responseEvents, EventResponse{
			Event:                evt,
			GuestCount:           guestCount,
			InventoryConsumed:    float64(int(consumption)), // round to nearest int %
			BudgetSpent:          budgetSpent,
			TotalBudget:          totalBudget,
			DaysUntilEvent:       daysUntil,
			PendingActions:       pendingActions,
			PendingActionDetails: pendingActionDetails,
		})
	}

	// 2. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(responseEvents); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 1*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Events Fetched Successfully",
		"events":  responseEvents,
	})
}

func (m *Repository) GetEvent(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	var event models.Event
	cacheKey := fmt.Sprintf("events:id:%s", id)
	ctx := context.Background()

	// Define response struct locally for the handler
	type EventResponse struct {
		models.Event
		GuestCount           int64    `json:"guestCount"`
		InventoryConsumed    float64  `json:"inventoryConsumed"`
		BudgetSpent          float64  `json:"budgetSpent"`
		TotalBudget          *float64 `json:"totalBudget"`
		DaysUntilEvent       int      `json:"daysUntilEvent"`
		PendingActions       int      `json:"pendingActions"`
		PendingActionDetails []string `json:"pendingActionDetails"`
	}
	var resEvent EventResponse

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &resEvent); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
					"message": "Event Fetched (Cached)",
					"event":   resEvent,
				})
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	if err := m.DB.Where("id = ?", id).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	// Fetch Agent Name
	var agent models.User
	agentName := "Agent"
	if err := m.DB.Where("id = ?", event.AgentID).First(&agent).Error; err == nil {
		agentName = agent.Name
	}
	event.Organizer = agentName

	var guestCount int64
	m.DB.Model(&models.Guest{}).Where("event_id = ?", event.ID).Count(&guestCount)

	var allocatedCount int64
	m.DB.Model(&models.GuestAllocation{}).Where("event_id = ?", event.ID).Count(&allocatedCount)

	var consumption float64 = 0.0
	if len(event.RoomsInventory) > 0 {
		var totalRooms int = 0
		var rooms []RoomsInventoryItem
		json.Unmarshal(event.RoomsInventory, &rooms)
		for _, rp := range rooms {
			totalRooms += rp.Available // Or rp.Total
		}
		if totalRooms > 0 {
			consumption = (float64(allocatedCount) / float64(totalRooms)) * 100
		}
	}

	// Calculate Days Until Event
	daysUntil := int(time.Until(event.StartDate).Hours() / 24)

	// Calculate Pending Actions (e.g. guests without allocations)
	var pendingGuests int64
	if event.Status != "locked" && event.Status != "finalized" {
		m.DB.Model(&models.Guest{}).
			Joins("LEFT JOIN guest_allocations ON guests.id = guest_allocations.guest_id").
			Where("guests.event_id = ? AND guest_allocations.id IS NULL", event.ID).
			Count(&pendingGuests)
	}

	pendingActions := int(pendingGuests)
	var pendingActionDetails []string
	if pendingGuests > 0 {
		pendingActionDetails = append(pendingActionDetails, fmt.Sprintf("%d guest(s) need room allocation", pendingGuests))
	}
	if event.HeadGuestID == uuid.Nil {
		pendingActions++
		pendingActionDetails = append(pendingActionDetails, "Assign a Head Guest")
	}

	// Calculate Budget Spent (Sum of LockedPrice in Allocations)
	var budgetSpent float64
	m.DB.Model(&models.GuestAllocation{}).Where("event_id = ?", event.ID).Select("COALESCE(SUM(locked_price), 0)").Scan(&budgetSpent)

	resEvent = EventResponse{
		Event:                event,
		GuestCount:           guestCount,
		InventoryConsumed:    float64(int(consumption)),
		BudgetSpent:          budgetSpent,
		TotalBudget:          event.Budget,
		DaysUntilEvent:       daysUntil,
		PendingActions:       pendingActions,
		PendingActionDetails: pendingActionDetails,
	}

	// 2. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(resEvent); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 1*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Event Fetched",
		"event":   resEvent,
	})
}

type CreateEventRequest struct {
	Name           string               `json:"name"` // Added Name field
	HotelID        string               `json:"hotelId"`
	Location       string               `json:"location"`
	StartDate      string               `json:"startDate"`
	EndDate        string               `json:"endDate"`
	Budget         *float64             `json:"budget"`
	RoomsInventory []RoomsInventoryItem `json:"roomsInventory"`
}

func (m *Repository) CreateEvent(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	agentID, ok := userID.(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid User ID type")
	}

	var req CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Basic Validation
	if req.Location == "" || req.StartDate == "" || req.EndDate == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	if req.Budget == nil || *req.Budget <= 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Budget must be greater than 0")
	}

	// Try parsing dates. Frontend sends YYYY-MM-DD usually for date inputs.
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, req.StartDate)
	if err != nil {
		// Try RFC3339 as fallback
		startDate, err = time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid StartDate format. Expected YYYY-MM-DD")
		}
	}

	endDate, err := time.Parse(layout, req.EndDate)
	if err != nil {
		endDate, err = time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid EndDate format. Expected YYYY-MM-DD")
		}
	}

	// Logical Date Validation
	now := time.Now().Truncate(24 * time.Hour)
	eventStart := startDate.Truncate(24 * time.Hour)
	eventEnd := endDate.Truncate(24 * time.Hour)

	if eventStart.Before(now) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Start Date cannot be in the past")
	}
	if eventEnd.Before(eventStart) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "End Date must be after or equal to Start Date")
	}

	// JSONB handling for RoomsInventory
	for i := range req.RoomsInventory {
		req.RoomsInventory[i].Total = req.RoomsInventory[i].Available
	}
	roomsJSON, _ := json.Marshal(req.RoomsInventory)

	event := models.Event{
		ID:             uuid.New(),
		AgentID:        agentID,
		Name:           req.Name, // Fix: Assign Name
		HotelID:        req.HotelID,
		Location:       req.Location,
		StartDate:      startDate,
		EndDate:        endDate,
		Budget:         req.Budget,
		Status:         "active",
		RoomsInventory: datatypes.JSON(roomsJSON),
	}

	if err := m.DB.Create(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create event")
	}

	// Invalidate agent events list
	utils.Invalidate(context.Background(), fmt.Sprintf("events:agent:%s", agentID.String()))

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Event Created Successfully",
		"event":   event,
	})
}

func (m *Repository) GetMetrics(c *fiber.Ctx) error {
	// TODO: Get metrics from store
	// metrics, err := m.DB.GetMetrics()
	// if err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to fetch metrics")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Get Metrics Endpoint",
		"metrics": []interface{}{},
	})
}

func (m *Repository) GetEventVenues(c *fiber.Ctx) error {
	// Get event ID from path parameter
	id := c.Params("id")

	// TODO: Get venues for event
	// venues, err := m.DB.GetVenuesByEventID(id)
	// if err != nil {
	//     return utils.InternalErrorResponse(c, "Failed to fetch venues")
	// }

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Get Event Venues Endpoint",
		"eventId": id,
		"venues":  []interface{}{},
	})
}

type UpdateEventRequest struct {
	Name           string                `json:"name"`
	HotelID        string                `json:"hotelId"`
	Location       string                `json:"location"`
	StartDate      string                `json:"startDate"`
	EndDate        string                `json:"endDate"`
	Budget         *float64              `json:"budget"`
	BudgetSpent    *float64              `json:"budgetSpent"`
	RoomsInventory *[]RoomsInventoryItem `json:"roomsInventory"`
}

func (m *Repository) UpdateEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	var event models.Event
	if err := m.DB.Where("id = ?", id).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.HotelID != "" {
		updates["hotel_id"] = req.HotelID
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.StartDate != "" {
		// Parse date
		layout := "2006-01-02"
		if t, err := time.Parse(layout, req.StartDate); err == nil {
			updates["start_date"] = t
		}
	}
	if req.EndDate != "" {
		layout := "2006-01-02"
		if t, err := time.Parse(layout, req.EndDate); err == nil {
			updates["end_date"] = t
		}
	}
	if req.Budget != nil {
		updates["budget"] = *req.Budget
	}
	if req.BudgetSpent != nil {
		updates["budget_spent"] = *req.BudgetSpent
	}

	if req.RoomsInventory != nil {
		roomsJSON, _ := json.Marshal(*req.RoomsInventory)
		updates["rooms_inventory"] = datatypes.JSON(roomsJSON)
	}

	if err := m.DB.Model(&event).Updates(updates).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update event")
	}

	// Invalidate cache
	utils.Invalidate(context.Background(),
		fmt.Sprintf("events:id:%s", id),
		fmt.Sprintf("events:agent:%s", event.AgentID.String()),
	)

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Event Updated Successfully",
		"event":   event,
	})
}

func (m *Repository) DeleteEvent(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// Fetch event first to get AgentID for cache invalidation
	var event models.Event
	if err := m.DB.Where("id = ?", id).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	// Transactional delete using Cascade if configured, or manual
	tx := m.DB.Begin()

	// Delete Allocations
	if err := tx.Where("event_id = ?", id).Delete(&models.GuestAllocation{}).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete allocations")
	}

	// Delete Guests
	if err := tx.Where("event_id = ?", id).Delete(&models.Guest{}).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete guests")
	}

	// Delete Negotiation Rounds explicitly (using subquery on sessions)
	if err := tx.Exec("DELETE FROM negotiation_rounds WHERE session_id IN (SELECT id FROM negotiation_sessions WHERE event_id = ?)", id).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete negotiation rounds")
	}

	// Delete Negotiation Sessions
	if err := tx.Exec("DELETE FROM negotiation_sessions WHERE event_id = ?", id).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete negotiation sessions")
	}

	// Delete Cart Items
	if err := tx.Where("event_id = ?", id).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete cart items")
	}

	// Delete Head Guest
	if event.HeadGuestID != uuid.Nil {
		if err := tx.Where("id = ?", event.HeadGuestID).Delete(&models.User{}).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete head guest")
		}
	}

	// Delete Event
	if err := tx.Where("id = ?", id).Delete(&models.Event{}).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete event")
	}

	tx.Commit()

	// Invalidate cache
	utils.Invalidate(context.Background(),
		fmt.Sprintf("events:id:%s", id),
		fmt.Sprintf("events:agent:%s", event.AgentID.String()),
	)

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Event Deleted Successfully",
		"id":      id,
	})
}

type AssignHeadGuestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Age   int    `json:"age"`
}

func (m *Repository) AssignHeadGuest(c *fiber.Ctx) error {
	id := c.Params("id")
	var req AssignHeadGuestRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Email == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Start a transaction
	tx := m.DB.Begin()
	if tx.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	// 1. Create or Find User
	// Create the Head Guest User account
	// TODO: Send email to head guest to set their password.
	// For now, setting a default password or handling this logic is required since PasswordHash is Not Null.
	// We'll use a placeholder hash for "ChangeMe123!"
	// defaultHash := "$2a$10$3QxDjD1ylg.6T4x.5.6.7.8.9.0.1.2.3.4.5.6.7.8.9.0.1.2" // Example hash or generate real one
	// Actually, let's use a dummy hash or generated one to avoid empty string error.

	// Better: Generate a random password and hash it, maybe print it or just store it.
	// simpler: just put a valid bcrypt hash to satisfy constraint.

	var user models.User
	var tempPassword string

	// Check if a user with this email already exists
	if err := tx.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// If user does not exist, create a new one
		if err == gorm.ErrRecordNotFound {
			log.Printf("👤 Creating new head guest user: %s", req.Email)
			// Create new user
			tempPassword = utils.GenerateTempPassword()
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
			if err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password")
			}

			user = models.User{
				ID:           uuid.New(),
				Email:        req.Email,
				Role:         "head_guest",
				Name:         req.Name,
				Phone:        req.Phone,
				PasswordHash: string(hashedPassword),
			}

			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create head guest user: "+err.Error())
			}
		} else {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to query user: "+err.Error())
		}
	} else {
		log.Printf("👤 User already exists: %s", req.Email)
		// If user exists, update their details if necessary and ensure role is head_guest
		if user.Role != "head_guest" {
			user.Role = "head_guest"
			if err := tx.Model(&user).Update("role", "head_guest").Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update existing user's role: "+err.Error())
			}
		}
		// Update other fields if provided in the request
		if req.Name != "" && user.Name != req.Name {
			if err := tx.Model(&user).Update("name", req.Name).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update existing user's name: "+err.Error())
			}
		}
		if req.Phone != "" && user.Phone != req.Phone {
			if err := tx.Model(&user).Update("phone", req.Phone).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update existing user's phone: "+err.Error())
			}
		}
	}

	// 2. Update Event
	// We cast string id to UUID or let GORM handle it? GORM usually handles string -> UUID if model is defined correctly.
	// However, it's safer to check if event exists first.
	var event models.Event
	if err := tx.Where("id = ?", id).First(&event).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	if err := tx.Model(&event).Update("head_guest_id", user.ID).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update event")
	}

	// 3. Create Guest Record (optional, but requested in previous conversations)
	// We ensure the Guest ID matches the User ID so portal lookups work
	guest := models.Guest{
		ID:            user.ID, // Link User ID to Guest ID
		EventID:       event.ID,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Age:           req.Age,
		Type:          "Adult",
		FamilyID:      uuid.New(),
		ArrivalDate:   event.StartDate,
		DepartureDate: event.EndDate,
	}
	if err := tx.Create(&guest).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create guest record")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	// Invalidate cache
	utils.Invalidate(context.Background(),
		fmt.Sprintf("events:id:%s", id),
		fmt.Sprintf("events:agent:%s", event.AgentID.String()),
	)

	// Send Email with Credentials if new user
	if tempPassword != "" {
		subject := fmt.Sprintf("Head Guest Access - %s", event.Name)
		body := fmt.Sprintf(`
			<h1>Welcome to %s!</h1>
			<p>You have been assigned as the Head Guest.</p>
			<p><strong>Login Details:</strong></p>
			<ul>
				<li>Email: %s</li>
				<li>Password: %s</li>
			</ul>
			<p>Please login to manage the event.</p>
		`, event.Name, user.Email, tempPassword)

		task, err := queue.NewEmailTask(user.Email, subject, body)
		if err == nil {
			if m.QueueClient != nil {
				if _, err := m.QueueClient.Enqueue(task); err != nil {
					log.Printf("❌ Failed to enqueue task: %v", err)
				} else {
					log.Printf("📧 Queued credential email for %s", user.Email)
				}
			} else {
				log.Println("❌ QueueClient is nil!")
			}
		} else {
			log.Printf("❌ Failed to create email task: %v", err)
		}
	} else {
		log.Printf("ℹ️ Skipping email for existing user %s (no temp password generated)", user.Email)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Head Guest Assigned Successfully. Credentials sent via email.",
		"user":    user,
	})
}
