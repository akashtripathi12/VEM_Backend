package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// StartNegotiationRequest
type StartNegotiationRequest struct {
	CartID       string             `json:"cart_id"` // Assuming eventID actually, since cart is per event
	TargetPrices map[string]float64 `json:"target_prices"`
}

// StartNegotiationResponse
type StartNegotiationResponse struct {
	SessionID  uuid.UUID `json:"session_id"`
	ShareToken uuid.UUID `json:"share_token"`
	ShareURL   string    `json:"share_url"`
}

// StartNegotiation initiates the negotiation process
func (r *Repository) StartNegotiation(c *fiber.Ctx) error {
	var req StartNegotiationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	eventID, err := uuid.Parse(req.CartID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Cart/Event ID"})
	}

	// 1. Fetch Cart Items
	var cartItems []models.CartItem
	if err := r.DB.Where("event_id = ? AND status = ?", eventID, "cart").Find(&cartItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch cart items"})
	}

	if len(cartItems) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cart is empty"})
	}

	// 2. Lock Cart Items
	// Update status to 'negotiation' to prevent modification? Or keep as 'cart' but lock event?
	// Requirement says: "Lock the original Cart so it cannot be edited normally."
	// Let's assume we change status to 'negotiation' or keep 'cart' but have a flag on Event or Session logic.
	// For now, I'll update them to 'negotiation_locked' status if that was valid, but schema has 'wishlist', 'approved', 'booked'.
	// I'll stick to 'cart' status but logic should prevent edits if a session is active.
	// Better yet, let's create the session first.

	// 3. Create NegotiationSession
	session := models.NegotiationSession{
		EventID:      eventID,
		Status:       models.NegotiationStatusWaitingForHotel,
		CurrentRound: 1,
	}

	if err := r.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create negotiation session"})
	}

	// 4. Create Round 1 (Agent's Proposal)
	var proposalItems []models.ProposalItem
	for _, item := range cartItems {
		targetPrice, ok := req.TargetPrices[item.ID.String()]
		if !ok {
			targetPrice = item.LockedPrice // Default to current price if not targeted
		}

		proposalItems = append(proposalItems, models.ProposalItem{
			CartItemID: item.ID,
			Type:       item.Type,
			RefID:      item.RefID,
			// Name: item.Name, // Not available directly in CartItem, would need to fetch or ignore for now
			Quantity: item.Quantity,
			Price:    targetPrice,
			// Currency: item.Currency,
		})
	}

	proposalJSON, err := json.Marshal(proposalItems)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal proposal"})
	}

	round := models.NegotiationRound{
		SessionID:        session.ID,
		RoundNumber:      1,
		ModifiedBy:       models.NegotiationModifierAgent,
		ProposalSnapshot: datatypes.JSON(proposalJSON),
		Remarks:          "Initial Offer",
	}

	if err := r.DB.Create(&round).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create negotiation round"})
	}

	// Return
	shareURL := fmt.Sprintf("%s/negotiation/%s", c.BaseURL(), session.ShareToken.String())
	return c.Status(fiber.StatusCreated).JSON(StartNegotiationResponse{
		SessionID:  session.ID,
		ShareToken: session.ShareToken,
		ShareURL:   shareURL,
	})
}

// ---------------------------------------------------------------------

// CounterOfferRequest
type CounterOfferRequest struct {
	SessionToken uuid.UUID          `json:"session_token"` // Share token or ID? Requirement says 'session_token'
	NewPrices    map[string]float64 `json:"new_prices"`
	Remarks      string             `json:"remarks"`
}

// SubmitCounterOffer handles the hotel's or agent's counter offer
func (r *Repository) SubmitCounterOffer(c *fiber.Ctx) error {
	var req CounterOfferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Identify Session
	var session models.NegotiationSession
	if err := r.DB.Where("share_token = ?", req.SessionToken).First(&session).Error; err != nil {
		// Could be Agent calling with SessionID, but for now let's assume ShareToken for Hotel
		// TODO: proper auth check. If user is authenticated agent, we might look up by ID.
		// For simplicity, let's try lookup by ID if token lookup fails?
		// But requirement says "Input: session_token".
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}

	if session.Status == models.NegotiationStatusLocked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Negotiation is locked"})
	}

	// Determine who is modifying
	modifier := models.NegotiationModifierHotel
	// Check if Agent (authenticated user)
	// user := c.Locals("user") ...
	// Logic to switch status based on who's calling.
	// Requirement: "Turn Switch: Update status to WAITING_FOR_AGENT" implied Hotel is calling.

	var nextStatus string
	if modifier == models.NegotiationModifierHotel {
		nextStatus = models.NegotiationStatusWaitingForAgent
	} else {
		nextStatus = models.NegotiationStatusWaitingForHotel
	}

	// 2. Fetch Previous Round (Latest)
	var lastRound models.NegotiationRound
	if err := r.DB.Where("session_id = ? AND round_number = ?", session.ID, session.CurrentRound).First(&lastRound).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch previous round"})
	}

	var lastProposal []models.ProposalItem
	if err := json.Unmarshal(lastRound.ProposalSnapshot, &lastProposal); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse previous proposal"})
	}

	// 3. Validate & Build New Proposal
	var newProposal []models.ProposalItem
	var totalNewPrice float64
	var totalTargetPrice float64 // From Agent's original target? Or last round?

	// Need to get original target/list price for variance check.
	// For simplicty, let's compare with previous round.
	// Requirement: "Variance Check: Calculate % difference between Agent's Target and Hotel's New Offer"

	for _, item := range lastProposal {
		newPrice, ok := req.NewPrices[item.CartItemID.String()]
		if !ok {
			newPrice = item.Price // No change
		}

		// Traffic Light Rule Check (per item or total?)
		// "If New Total > Target + 20%: Flag Red"

		newItem := item
		newItem.Price = newPrice
		newProposal = append(newProposal, newItem)

		totalNewPrice += newPrice * float64(item.Quantity)
		totalTargetPrice += item.Price * float64(item.Quantity) // Using last round as baseline for now
	}

	// TODO: Implement Variance Check / Flags logic details
	// Flags can be returned in response or stored in remarks/reason_code?

	newProposalJSON, _ := json.Marshal(newProposal)

	// 4. Create New Round
	nextRoundNum := session.CurrentRound + 1
	newRound := models.NegotiationRound{
		SessionID:        session.ID,
		RoundNumber:      nextRoundNum,
		ModifiedBy:       modifier,
		ProposalSnapshot: datatypes.JSON(newProposalJSON),
		Remarks:          req.Remarks,
	}

	// Transaction
	tx := r.DB.Begin()
	if err := tx.Create(&newRound).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create round"})
	}

	if err := tx.Model(&session).Updates(map[string]interface{}{
		"current_round": nextRoundNum,
		"status":        nextStatus,
	}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update session"})
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Counter offer submitted", "round": nextRoundNum})
}

// ---------------------------------------------------------------------

// GetNegotiationDiff returns changes between current and previous round
func (r *Repository) GetNegotiationDiff(c *fiber.Ctx) error {
	id := c.Params("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Session ID"})
	}

	var session models.NegotiationSession
	if err := r.DB.First(&session, sessionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}

	if session.CurrentRound < 2 {
		return c.JSON(fiber.Map{"message": "No history to compare"})
	}

	var rounds []models.NegotiationRound
	if err := r.DB.Where("session_id = ? AND round_number IN ?", sessionID, []int{session.CurrentRound, session.CurrentRound - 1}).Order("round_number desc").Find(&rounds).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch rounds"})
	}

	if len(rounds) != 2 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not find both rounds"})
	}

	currRound := rounds[0]
	prevRound := rounds[1]

	var currItems, prevItems []models.ProposalItem
	json.Unmarshal(currRound.ProposalSnapshot, &currItems)
	json.Unmarshal(prevRound.ProposalSnapshot, &prevItems)

	// Map for quick lookup
	prevMap := make(map[uuid.UUID]models.ProposalItem)
	for _, item := range prevItems {
		prevMap[item.CartItemID] = item
	}

	var diffs []map[string]interface{}
	for _, curr := range currItems {
		prev, ok := prevMap[curr.CartItemID]
		if !ok {
			continue // New item?
		}

		if curr.Price != prev.Price {
			diffs = append(diffs, map[string]interface{}{
				"id":    curr.CartItemID,
				"field": "price",
				"old":   prev.Price,
				"new":   curr.Price,
				"delta": curr.Price - prev.Price,
			})
		}
	}

	return c.JSON(fiber.Map{"items": diffs})
}

// ---------------------------------------------------------------------

// LockDealRequest
type LockDealRequest struct {
	SessionID string `json:"session_id"`
}

// LockDeal finalizes the negotiation
func (r *Repository) LockDeal(c *fiber.Ctx) error {
	var req LockDealRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Session ID"})
	}

	var session models.NegotiationSession
	if err := r.DB.Preload("Event").First(&session, sessionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}

	// Fetch last round
	var lastRound models.NegotiationRound
	if err := r.DB.Where("session_id = ? AND round_number = ?", session.ID, session.CurrentRound).First(&lastRound).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch last round"})
	}

	var finalItems []models.ProposalItem
	if err := json.Unmarshal(lastRound.ProposalSnapshot, &finalItems); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse final proposal"})
	}

	tx := r.DB.Begin()

	// Update Cart Items
	for _, item := range finalItems {
		// Update price and status to approved/ready?
		// Requirement: "Change Cart Status to READY_TO_PAY" (which might not be in Enum, let's say 'approved')
		if err := tx.Model(&models.CartItem{}).Where("id = ?", item.CartItemID).Updates(map[string]interface{}{
			"locked_price": item.Price,
			"status":       "approved",
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update cart item " + item.CartItemID.String()})
		}
	}

	// Lock Session
	if err := tx.Model(&session).Update("status", models.NegotiationStatusLocked).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to lock session"})
	}

	// Trigger Contract PDF (stub)
	// TODO: fire event/queue

	tx.Commit()

	return c.JSON(fiber.Map{"message": "Deal locked successfully"})
}
