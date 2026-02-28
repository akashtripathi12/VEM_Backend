package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Request DTOs ---

type AddToCartRequest struct {
	Type       string  `json:"type"`         // 'room', 'banquet', 'catering', 'flight', 'transfer'
	RefID      string  `json:"refId"`        // ID of the referenced item
	Quantity   int     `json:"quantity"`     // Number of units (default: 1)
	TaxAndFees float64 `json:"tax_and_fees"` // Taxes and fees
	Notes      string  `json:"notes"`        // Optional notes
	Status     string  `json:"status"`       // Optional status (default: 'wishlist')
}

type UpdateCartItemRequest struct {
	Quantity    *int     `json:"quantity,omitempty"`     // Optional quantity update
	Notes       *string  `json:"notes,omitempty"`        // Optional notes update
	Status      *string  `json:"status,omitempty"`       // Optional status update
	LockedPrice *float64 `json:"locked_price,omitempty"` // Optional locked price update
}

// AddToCart adds an item to the event cart/wishlist
func (m *Repository) AddToCart(c *fiber.Ctx) error {
	eventID := c.Params("id")
	userID := c.Locals("userID")
	if userID == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	currentUserID, ok := userID.(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid User ID type")
	}

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// Guard: Check Event Status (New Lifecycle)
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Parse request
	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Type == "" || req.RefID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing required fields: type and refId")
	}

	// Default quantity to 1 if not provided
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	// Default status if not provided
	if req.Status == "" {
		req.Status = "wishlist"
	}

	// Determine parent_hotel_id and fetch locked price based on type
	var parentHotelID *string
	var lockedPrice float64
	var flightBookingID *uuid.UUID
	var transferBookingID *uuid.UUID

	switch req.Type {
	case "room":
		var room models.RoomOffer
		log.Printf("🔍 [CART] Searching for RoomOffer ID: %s\n", req.RefID)
		if err := m.DB.Where("id = ?", req.RefID).First(&room).Error; err != nil {
			log.Printf("❌ [CART] RoomOffer NOT FOUND: %s (Error: %v)\n", req.RefID, err)
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Room offer not found")
		}
		parentHotelID = &room.HotelID
		lockedPrice = room.TotalFare

	case "banquet":
		var banquet models.BanquetHall
		if err := m.DB.Where("id = ?", req.RefID).First(&banquet).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Banquet hall not found")
		}
		parentHotelID = &banquet.HotelID
		lockedPrice = banquet.PricePerDay

	case "catering":
		var catering models.CateringMenu
		if err := m.DB.Where("id = ?", req.RefID).First(&catering).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Catering menu not found")
		}
		parentHotelID = &catering.HotelID
		lockedPrice = catering.PricePerPlate

	case "hotel":
		var hotel models.Hotel
		log.Printf("🔍 [CART] Searching for Hotel Code: %s\n", req.RefID)
		if err := m.DB.Where("hotel_code = ?", req.RefID).First(&hotel).Error; err != nil {
			log.Printf("❌ [CART] Hotel NOT FOUND: %s (Error: %v)\n", req.RefID, err)
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Hotel not found")
		}
		parentHotelID = &hotel.ID
		lockedPrice = 0

	case "flight":
		var flight models.Flight
		if err := m.DB.Where("id = ?", req.RefID).First(&flight).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Flight not found")
		}

		// Check availability
		if flight.AvailableSeats < req.Quantity {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Not enough seats available. Available: %d, Requested: %d", flight.AvailableSeats, req.Quantity))
		}

		parentHotelID = nil
		lockedPrice = flight.BasePrice

		// Create Flight Booking
		eventUUID, _ := uuid.Parse(eventID)
		flightBooking := models.FlightBooking{
			FlightID:    flight.ID,
			EventID:     eventUUID,
			SeatsBooked: req.Quantity,
			PriceLocked: flight.BasePrice,
			Status:      "pending",
			BookedBy:    currentUserID,
		}

		// Transaction to update inventory and booking
		tx := m.DB.Begin()

		// Decrement persistence
		flight.AvailableSeats -= req.Quantity
		if err := tx.Save(&flight).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update flight inventory")
		}

		if err := tx.Create(&flightBooking).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create flight booking")
		}

		if err := tx.Commit().Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit booking")
		}

		flightBookingID = &flightBooking.ID

	case "transfer":
		var transfer models.Transfer
		if err := m.DB.Where("id = ?", req.RefID).First(&transfer).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Transfer not found")
		}
		// Check availability
		if transfer.AvailableCount < req.Quantity {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Not enough cabs available. Available: %d, Requested: %d", transfer.AvailableCount, req.Quantity))
		}
		parentHotelID = nil
		lockedPrice = transfer.BasePricePerCab

		// Create Transfer Booking
		eventUUID, _ := uuid.Parse(eventID)
		transferBooking := models.TransferBooking{
			TransferID:     transfer.ID,
			EventID:        eventUUID,
			CabsBooked:     req.Quantity,
			PriceLocked:    transfer.BasePricePerCab,
			PickupLocation: "To be decided", // Default
			DropLocation:   "To be decided", // Default
			Status:         "pending",
			BookedBy:       currentUserID,
		}

		// Transaction
		tx := m.DB.Begin()

		// Decrement persistence
		transfer.AvailableCount -= req.Quantity
		if err := tx.Save(&transfer).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update transfer inventory")
		}

		if err := tx.Create(&transferBooking).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create transfer booking")
		}

		if err := tx.Commit().Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit booking")
		}

		transferBookingID = &transferBooking.ID

	default:
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid type. Must be: hotel, room, banquet, catering, flight, or transfer")
	}

	// Create cart item — but first check for duplicate (upsert by quantity for rooms/banquets/catering/hotel)
	eventUUID, _ := uuid.Parse(eventID)

	if req.Type != "flight" && req.Type != "transfer" {
		var existing models.CartItem
		result := m.DB.Where("event_id = ? AND type = ? AND ref_id = ?", eventUUID, req.Type, req.RefID).First(&existing)
		if result.Error == nil {
			// Item already in cart — just increment quantity
			newQty := existing.Quantity + req.Quantity
			if err := m.DB.Model(&existing).Update("quantity", newQty).Error; err != nil {
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update cart item quantity")
			}
			existing.Quantity = newQty
			utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))
			return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
				"message": "Cart item quantity updated",
				"item":    existing,
			})
		}
	}

	cartItem := models.CartItem{
		EventID:           eventUUID,
		Type:              req.Type,
		RefID:             req.RefID,
		ParentHotelID:     parentHotelID,
		FlightBookingID:   flightBookingID,
		TransferBookingID: transferBookingID,
		Status:            req.Status,
		Quantity:          req.Quantity,
		LockedPrice:       lockedPrice,
		TaxAndFees:        req.TaxAndFees,
		Notes:             req.Notes,
		AddedBy:           currentUserID,
	}

	if err := m.DB.Create(&cartItem).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add item to cart")
	}

	// Invalidate Cache
	utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Item added to cart successfully",
		"item":    cartItem,
	})
}

// BulkAddToCart adds multiple items to the event cart/wishlist in a single transaction
func (m *Repository) BulkAddToCart(c *fiber.Ctx) error {
	eventID := c.Params("id")
	userID := c.Locals("userID")
	if userID == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	currentUserID, ok := userID.(uuid.UUID)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid User ID type")
	}

	// Validate event ID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// Guard: Check Event Status
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Parse request array
	var reqs []AddToCartRequest
	if err := c.BodyParser(&reqs); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(reqs) == 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "No items provided")
	}

	var cartItems []models.CartItem
	tx := m.DB.Begin()

	// Process each request
	for _, req := range reqs {
		// Default quantity & status
		if req.Quantity <= 0 {
			req.Quantity = 1
		}
		if req.Status == "" {
			req.Status = "wishlist"
		}

		var parentHotelID *string
		var lockedPrice float64

		// Note: To keep things simple and safe, we only allow rooms, banquets, catering, and hotels
		// in the bulk hotel selection. Flights and transfers are complex and still need individual inventory updates
		switch req.Type {
		case "room":
			var room models.RoomOffer
			if err := tx.Where("id = ?", req.RefID).First(&room).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusNotFound, fmt.Sprintf("Room offer not found: %s", req.RefID))
			}
			parentHotelID = &room.HotelID
			lockedPrice = room.TotalFare
		case "banquet":
			var banquet models.BanquetHall
			if err := tx.Where("id = ?", req.RefID).First(&banquet).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusNotFound, "Banquet hall not found")
			}
			parentHotelID = &banquet.HotelID
			lockedPrice = banquet.PricePerDay
		case "catering":
			var catering models.CateringMenu
			if err := tx.Where("id = ?", req.RefID).First(&catering).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusNotFound, "Catering menu not found")
			}
			parentHotelID = &catering.HotelID
			lockedPrice = catering.PricePerPlate
		case "hotel":
			var hotel models.Hotel
			if err := tx.Where("hotel_code = ?", req.RefID).First(&hotel).Error; err != nil {
				tx.Rollback()
				return utils.ErrorResponse(c, fiber.StatusNotFound, "Hotel not found")
			}
			parentHotelID = &hotel.ID
			lockedPrice = 0
		default:
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Bulk add only supports hotel, room, banquet, and catering types")
		}

		cartItems = append(cartItems, models.CartItem{
			EventID:       eventUUID,
			Type:          req.Type,
			RefID:         req.RefID,
			ParentHotelID: parentHotelID,
			Status:        req.Status,
			Quantity:      req.Quantity,
			LockedPrice:   lockedPrice,
			TaxAndFees:    req.TaxAndFees,
			Notes:         req.Notes,
			AddedBy:       currentUserID,
		})
	}

	// Bulk Insert
	if len(cartItems) > 0 {
		if err := tx.Create(&cartItems).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add items to cart")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit items")
	}

	// Invalidate Cache
	utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Items added to cart successfully",
		"items":   cartItems,
	})
}

// GetEventCart retrieves all cart items for an event with hierarchical grouping
func (r *Repository) GetEventCart(c *fiber.Ctx) error {
	eventID := c.Params("id")
	status := c.Query("status") // Optional filter by status

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	var response models.CartResponse
	cacheKey := fmt.Sprintf("cart:event:%s", eventID)
	if status != "" {
		cacheKey += ":status:" + status
	}
	ctx := context.Background()

	// 1. Try to get from Redis
	if store.RDB != nil {
		cachedData, err := store.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				log.Printf("⚡ [REDIS] CACHE HIT: %s\n", cacheKey)
				return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
					"message": "Cart fetched successfully (Cached)",
					"cart":    response,
				})
			}
		} else {
			log.Printf("🔍 [REDIS] CACHE MISS: %s (Reason: %v)\n", cacheKey, err)
		}
	}

	// Fetch cart items
	var cartItems []models.CartItem
	query := r.DB.Where("event_id = ?", eventID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&cartItems).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch cart items")
	}

	// Build hierarchical response
	var err error
	response, err = r.buildHierarchicalCartResponse(eventID, cartItems)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to build cart response: "+err.Error())
	}

	// 2. Store in Redis
	if store.RDB != nil {
		if data, err := json.Marshal(response); err == nil {
			store.RDB.Set(ctx, cacheKey, data, 1*time.Hour)
			log.Printf("💾 [REDIS] CACHE SET: %s\n", cacheKey)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Cart fetched successfully",
		"cart":    response,
	})
}

// UpdateCartItem updates a cart item's quantity, notes, or status
func (m *Repository) UpdateCartItem(c *fiber.Ctx) error {
	eventID := c.Params("id")
	cartItemID := c.Params("cartItemId")

	// Validate IDs
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}
	if _, err := uuid.Parse(cartItemID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Cart Item ID")
	}

	// Guard: Check Event Status
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Parse request
	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Find cart item
	var cartItem models.CartItem
	if err := m.DB.Where("id = ? AND event_id = ?", cartItemID, eventID).First(&cartItem).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Cart item not found")
	}

	// Build updates map
	updates := map[string]interface{}{}
	if req.Quantity != nil && *req.Quantity > 0 {
		updates["quantity"] = *req.Quantity
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.LockedPrice != nil {
		updates["locked_price"] = *req.LockedPrice
	}

	// Apply updates
	if len(updates) > 0 {
		// Synced inventory update for Flights/Transfers
		if req.Quantity != nil && *req.Quantity > 0 && *req.Quantity != cartItem.Quantity {
			delta := *req.Quantity - cartItem.Quantity

			// Start transaction for inventory sync
			tx := m.DB.Begin()

			if cartItem.Type == "flight" && cartItem.FlightBookingID != nil {
				var flight models.Flight
				var booking models.FlightBooking

				if err := tx.First(&booking, cartItem.FlightBookingID).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Flight booking not found")
				}
				if err := tx.First(&flight, booking.FlightID).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Flight not found")
				}

				// Check availability if increasing quantity
				if delta > 0 && flight.AvailableSeats < delta {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Not enough seats available. Available: %d, Requested Additional: %d", flight.AvailableSeats, delta))
				}

				// Update inventory
				if err := tx.Model(&flight).Update("available_seats", gorm.Expr("available_seats - ?", delta)).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update flight inventory")
				}

				// Update booking quantity
				if err := tx.Model(&booking).Update("seats_booked", *req.Quantity).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update booking quantity")
				}

			} else if cartItem.Type == "transfer" && cartItem.TransferBookingID != nil {
				var transfer models.Transfer
				var booking models.TransferBooking

				if err := tx.First(&booking, cartItem.TransferBookingID).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Transfer booking not found")
				}
				if err := tx.First(&transfer, booking.TransferID).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Transfer not found")
				}

				// Check availability if increasing quantity
				if delta > 0 && transfer.AvailableCount < delta {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Not enough cabs available. Available: %d, Requested Additional: %d", transfer.AvailableCount, delta))
				}

				// Update inventory
				if err := tx.Model(&transfer).Update("available_count", gorm.Expr("available_count - ?", delta)).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update transfer inventory")
				}

				// Update booking quantity
				if err := tx.Model(&booking).Update("cabs_booked", *req.Quantity).Error; err != nil {
					tx.Rollback()
					return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update booking quantity")
				}
			}

			// Commit inventory changes
			if err := tx.Commit().Error; err != nil {
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit inventory updates")
			}
		}

		// Update cart item (including status/notes if present)
		if err := m.DB.Model(&cartItem).Updates(updates).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update cart item")
		}

		// Invalidate Cache
		utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))
	}

	// Fetch updated item
	m.DB.Where("id = ?", cartItemID).First(&cartItem)

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Cart item updated successfully",
		"item":    cartItem,
	})
}

// RemoveFromCart removes a cart item
func (m *Repository) RemoveFromCart(c *fiber.Ctx) error {
	eventID := c.Params("id")
	cartItemID := c.Params("cartItemId")

	// Validate IDs
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}
	if _, err := uuid.Parse(cartItemID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Cart Item ID")
	}

	// Guard: Check Event Status
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Find item first to get booking IDs
	var cartItem models.CartItem
	if err := m.DB.Where("id = ? AND event_id = ?", cartItemID, eventID).First(&cartItem).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Cart item not found")
	}

	itemType := cartItem.Type
	var flightBookingID *uuid.UUID
	if cartItem.FlightBookingID != nil {
		flightBookingID = cartItem.FlightBookingID
	}
	var transferBookingID *uuid.UUID
	if cartItem.TransferBookingID != nil {
		transferBookingID = cartItem.TransferBookingID
	}

	// Start transaction
	tx := m.DB.Begin()

	// 1. Delete cart item FIRST to clear any foreign key dependencies
	if err := tx.Delete(&cartItem).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to remove cart item: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to remove cart item")
	}

	// 2. Clear out the Booking and Restore Inventory
	if itemType == "flight" && flightBookingID != nil {
		var booking models.FlightBooking
		if err := tx.First(&booking, flightBookingID).Error; err == nil {
			// Restore seats
			var flight models.Flight
			if err := tx.First(&flight, booking.FlightID).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to find flight: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to find flight")
			}
			flight.AvailableSeats += booking.SeatsBooked
			if err := tx.Save(&flight).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to restore flight seats: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to restore flight seats")
			}
			// Delete booking
			if err := tx.Delete(&booking).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to delete flight booking: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete flight booking")
			}
		}
	} else if itemType == "transfer" && transferBookingID != nil {
		var booking models.TransferBooking
		if err := tx.First(&booking, transferBookingID).Error; err == nil {
			// Restore count
			var transfer models.Transfer
			if err := tx.First(&transfer, booking.TransferID).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to find transfer: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to find transfer")
			}
			transfer.AvailableCount += booking.CabsBooked
			if err := tx.Save(&transfer).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to restore transfer cabs: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to restore transfer cabs")
			}
			// Delete booking
			if err := tx.Delete(&booking).Error; err != nil {
				tx.Rollback()
				log.Printf("ERROR: Failed to delete transfer booking: %v", err)
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete transfer booking")
			}
		}
	}

	// 3. Reset Negotiations (Delete sessions & rounds for this event)
	if err := tx.Exec("DELETE FROM negotiation_rounds WHERE session_id IN (SELECT id FROM negotiation_sessions WHERE event_id = ?)", eventID).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to clear negotiation rounds: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to clear negotiation rounds")
	}
	if err := tx.Exec("DELETE FROM negotiation_sessions WHERE event_id = ?", eventID).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to clear negotiation sessions: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to clear negotiation sessions")
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR: Failed to commit removal: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit removal")
	}

	// Invalidate Cache
	utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Cart item removed successfully",
	})
}

// RemoveHotelGroupFromCart completely removes a hotel and all its nested children
func (m *Repository) RemoveHotelGroupFromCart(c *fiber.Ctx) error {
	eventID := c.Params("id")
	hotelID := c.Params("hotelId") // The parent_hotel_id string

	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// Guard: Check Event Status
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Delete all cart items related to this event and having this hotel_id
	// This hits both the base 'hotel' type (with RefID = hotelId) and subitems (ParentHotelID = hotelId)
	result := m.DB.Where("event_id = ? AND (parent_hotel_id = ? OR (type = 'hotel' AND ref_id = ?))", eventID, hotelID, hotelID).
		Delete(&models.CartItem{})

	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to remove hotel group items")
	}

	// Reset Negotiations (Delete sessions & rounds for this event) since cart changed
	if err := m.DB.Exec("DELETE FROM negotiation_rounds WHERE session_id IN (SELECT id FROM negotiation_sessions WHERE event_id = ?)", eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to clear negotiation rounds")
	}
	if err := m.DB.Exec("DELETE FROM negotiation_sessions WHERE event_id = ?", eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to clear negotiation sessions")
	}

	// Invalidate Cache
	utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message":       "Hotel group removed successfully",
		"items_deleted": result.RowsAffected,
	})
}

// UpdateCartStatus approves all wishlist items (converts wishlist to approved cart)
func (m *Repository) UpdateCartStatus(c *fiber.Ctx) error {
	eventID := c.Params("id")

	// Validate event ID
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// Guard: Check Event Status
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}
	if event.Status == "finalized" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Event is finalized and locked")
	}

	// Update all wishlist items to approved
	result := m.DB.Model(&models.CartItem{}).
		Where("event_id = ? AND status = ?", eventID, "wishlist").
		Update("status", "approved")

	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to approve cart items")
	}

	// Invalidate Cache
	utils.Invalidate(context.Background(), fmt.Sprintf("cart:event:%s", eventID))

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message":       "Cart approved successfully",
		"items_updated": result.RowsAffected,
	})
}

// --- Helper Functions ---

// buildHierarchicalCartResponse constructs the hierarchical response grouped by hotel
func (m *Repository) buildHierarchicalCartResponse(eventID string, cartItems []models.CartItem) (models.CartResponse, error) {
	eventUUID, _ := uuid.Parse(eventID)
	response := models.CartResponse{
		EventID:   eventUUID,
		Hotels:    []models.HotelCartGroup{},
		Flights:   []models.CartItemDetail{},
		Transfers: []models.CartItemDetail{},
	}

	// Group items by hotel
	hotelGroups := make(map[string]*models.HotelCartGroup)

	// Collect all IDs for batch fetching
	var roomIDs, banquetIDs, cateringIDs, flightIDs, transferIDs []string
	var hotelIDs []string

	for _, item := range cartItems {
		if item.Type == "flight" {
			flightIDs = append(flightIDs, item.RefID)
		} else if item.Type == "transfer" {
			transferIDs = append(transferIDs, item.RefID)
		} else {
			// Group by hotel
			hotelID := ""
			if item.ParentHotelID != nil {
				hotelID = *item.ParentHotelID
			}

			if _, exists := hotelGroups[hotelID]; !exists {
				hotelGroups[hotelID] = &models.HotelCartGroup{
					Rooms:    []models.CartItemDetail{},
					Banquets: []models.CartItemDetail{},
					Catering: []models.CartItemDetail{},
				}
				if hotelID != "" {
					hotelIDs = append(hotelIDs, hotelID)
				}
			}

			// Collect IDs for batch fetch
			switch item.Type {
			case "room":
				roomIDs = append(roomIDs, item.RefID)
			case "banquet":
				banquetIDs = append(banquetIDs, item.RefID)
			case "catering":
				cateringIDs = append(cateringIDs, item.RefID)
			}
		}
	}

	// Batch fetch all details
	rooms := m.fetchRoomDetails(roomIDs)
	banquets := m.fetchBanquetDetails(banquetIDs)
	caterings := m.fetchCateringDetails(cateringIDs)
	flights := m.fetchFlightDetails(flightIDs)
	transfers := m.fetchTransferDetails(transferIDs)
	hotels := m.fetchHotelDetails(hotelIDs)

	// Map items to groups with details
	for _, item := range cartItems {
		if item.Type == "flight" {
			cartDetail := models.CartItemDetail{
				ID:          item.ID,
				Type:        item.Type,
				Status:      item.Status,
				Quantity:    item.Quantity,
				LockedPrice: item.LockedPrice,
				TaxAndFees:  item.TaxAndFees,
				RefID:       item.RefID,
				Notes:       item.Notes,
				CreatedAt:   item.CreatedAt,
			}
			if flightDetail, ok := flights[item.RefID]; ok {
				cartDetail.FlightDetails = flightDetail
			}
			response.Flights = append(response.Flights, cartDetail)
			continue
		}

		if item.Type == "transfer" {
			cartDetail := models.CartItemDetail{
				ID:          item.ID,
				Type:        item.Type,
				Status:      item.Status,
				Quantity:    item.Quantity,
				LockedPrice: item.LockedPrice,
				TaxAndFees:  item.TaxAndFees,
				RefID:       item.RefID,
				Notes:       item.Notes,
				CreatedAt:   item.CreatedAt,
			}
			if transferDetail, ok := transfers[item.RefID]; ok {
				cartDetail.TransferDetails = transferDetail
			}
			response.Transfers = append(response.Transfers, cartDetail)
			continue
		}

		hotelID := ""
		if item.ParentHotelID != nil {
			hotelID = *item.ParentHotelID
		}

		cartDetail := models.CartItemDetail{
			ID:          item.ID,
			Type:        item.Type,
			Status:      item.Status,
			Quantity:    item.Quantity,
			LockedPrice: item.LockedPrice,
			TaxAndFees:  item.TaxAndFees,
			RefID:       item.RefID,
			Notes:       item.Notes,
			CreatedAt:   item.CreatedAt,
		}

		switch item.Type {
		case "room":
			if roomDetail, ok := rooms[item.RefID]; ok {
				cartDetail.RoomDetails = roomDetail
			}
			hotelGroups[hotelID].Rooms = append(hotelGroups[hotelID].Rooms, cartDetail)

		case "banquet":
			if banquetDetail, ok := banquets[item.RefID]; ok {
				cartDetail.BanquetDetails = banquetDetail
			}
			hotelGroups[hotelID].Banquets = append(hotelGroups[hotelID].Banquets, cartDetail)

		case "catering":
			if cateringDetail, ok := caterings[item.RefID]; ok {
				cartDetail.CateringDetails = cateringDetail
			}
			hotelGroups[hotelID].Catering = append(hotelGroups[hotelID].Catering, cartDetail)

		case "hotel":
			if hotelDetail, ok := hotels[item.RefID]; ok {
				cartDetail.HotelDetails = hotelDetail
			}
			hotelGroups[hotelID].HotelWishlistItem = &cartDetail
		}
	}

	// Build final hotel groups with hotel details
	for hotelID, group := range hotelGroups {
		if hotelDetail, ok := hotels[hotelID]; ok {
			group.HotelDetails = hotelDetail
		}
		response.Hotels = append(response.Hotels, *group)
	}

	return response, nil
}

// fetchHotelDetails fetches hotel details for given hotel IDs
func (m *Repository) fetchHotelDetails(hotelIDs []string) map[string]models.Hotel {
	result := make(map[string]models.Hotel)
	if len(hotelIDs) == 0 {
		return result
	}

	var hotels []models.Hotel
	m.DB.Where("hotel_code IN ?", hotelIDs).Find(&hotels)

	for _, hotel := range hotels {
		result[hotel.ID] = hotel
	}
	return result
}

// fetchRoomDetails fetches room offer details for given room IDs
func (m *Repository) fetchRoomDetails(roomIDs []string) map[string]models.RoomOffer {
	result := make(map[string]models.RoomOffer)
	if len(roomIDs) == 0 {
		return result
	}

	var rooms []models.RoomOffer
	m.DB.Where("id IN ?", roomIDs).Find(&rooms)

	for _, room := range rooms {
		result[room.ID] = room
	}
	return result
}

// fetchBanquetDetails fetches banquet hall details for given banquet IDs
func (m *Repository) fetchBanquetDetails(banquetIDs []string) map[string]models.BanquetHall {
	result := make(map[string]models.BanquetHall)
	if len(banquetIDs) == 0 {
		return result
	}

	var banquets []models.BanquetHall
	m.DB.Where("id IN ?", banquetIDs).Find(&banquets)

	for _, banquet := range banquets {
		result[fmt.Sprintf("%d", banquet.ID)] = banquet
	}
	return result
}

// fetchCateringDetails fetches catering menu details for given catering IDs
func (m *Repository) fetchCateringDetails(cateringIDs []string) map[string]models.CateringMenu {
	result := make(map[string]models.CateringMenu)
	if len(cateringIDs) == 0 {
		return result
	}

	var caterings []models.CateringMenu
	m.DB.Where("id IN ?", cateringIDs).Find(&caterings)

	for _, catering := range caterings {
		result[fmt.Sprintf("%d", catering.ID)] = catering
	}
	return result
}

// fetchFlightDetails fetches flight details for given flight IDs
func (m *Repository) fetchFlightDetails(flightIDs []string) map[string]models.Flight {
	result := make(map[string]models.Flight)
	if len(flightIDs) == 0 {
		return result
	}

	var flights []models.Flight
	m.DB.Where("id IN ?", flightIDs).Find(&flights)

	for _, flight := range flights {
		result[flight.ID.String()] = flight
	}
	return result
}

// fetchTransferDetails fetches transfer details for given transfer IDs
func (m *Repository) fetchTransferDetails(transferIDs []string) map[string]models.Transfer {
	result := make(map[string]models.Transfer)
	if len(transferIDs) == 0 {
		return result
	}

	var transfers []models.Transfer
	m.DB.Where("id IN ?", transferIDs).Find(&transfers)

	for _, transfer := range transfers {
		result[transfer.ID.String()] = transfer
	}
	return result
}
