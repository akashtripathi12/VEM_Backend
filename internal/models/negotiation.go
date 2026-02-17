package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// NegotiationStatus enum
const (
	NegotiationStatusDraft           = "draft"
	NegotiationStatusWaitingForAgent = "waiting_for_agent"
	NegotiationStatusWaitingForHotel = "waiting_for_hotel"
	NegotiationStatusLocked          = "locked"
)

// NegotiationModifier enum
const (
	NegotiationModifierAgent = "agent"
	NegotiationModifierHotel = "hotel"
)

// NegotiationReason enum
const (
	NegotiationReasonBudgetConstraint = "budget_constraint"
	NegotiationReasonVolumeDiscount   = "volume_discount"
	NegotiationReasonCompetitorOffer  = "competitor_offer"
	NegotiationReasonOther            = "other"
)

// NegotiationSession tracks the state of a negotiation
type NegotiationSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID      uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Event        Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Status       string    `gorm:"type:negotiation_status;default:'draft';not null" json:"status"`
	ShareToken   uuid.UUID `gorm:"type:uuid;uniqueIndex;default:gen_random_uuid()" json:"share_token"`
	CurrentRound int       `gorm:"default:1;not null" json:"current_round"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	Rounds []NegotiationRound `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"rounds,omitempty"`
}

// NegotiationRound stores the snapshot of a specific turn in the negotiation
type NegotiationRound struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	RoundNumber      int            `gorm:"not null" json:"round_number"`
	ModifiedBy       string         `gorm:"type:negotiation_modifier;not null" json:"modified_by"` // 'agent' or 'hotel'
	ProposalSnapshot datatypes.JSON `gorm:"type:jsonb;not null" json:"proposal_snapshot"`          // Array of ProposalItem
	Remarks          string         `gorm:"type:text" json:"remarks"`
	ReasonCode       string         `gorm:"type:negotiation_reason" json:"reason_code"`
	CreatedAt        time.Time      `json:"created_at"`
}

// ProposalItem is a helper struct for the JSON snapshot
// It mirrors the necessary fields from CartItem for price negotiation
type ProposalItem struct {
	CartItemID uuid.UUID `json:"cart_item_id"`
	Type       string    `json:"type"`          // room, banquet, etc.
	RefID      string    `json:"ref_id"`        // to identify the item
	Name       string    `json:"name"`          // snapshot of item name
	Quantity   int       `json:"quantity"`
	Price      float64   `json:"price"` // The proposed price
	Currency   string    `json:"currency"`
	Inclusions []string  `json:"inclusions,omitempty"` // snapshot of inclusions
}
