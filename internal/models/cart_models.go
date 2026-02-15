package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartItem represents an item in the event cart/wishlist
// Uses polymorphic design: Type + RefID to reference different tables
type CartItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID       uuid.UUID `gorm:"type:uuid;index;not null" json:"event_id"`
	Event         Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Type          string    `gorm:"size:20;not null;index" json:"type"`             // 'room', 'banquet', 'catering', 'flight'
	RefID         string    `gorm:"size:255;not null" json:"ref_id"`                // ID of the referenced item
	ParentHotelID *string   `gorm:"size:50;index" json:"parent_hotel_id,omitempty"` // Hotel code for grouping (NULL for flights)
	Status        string    `gorm:"size:20;default:'wishlist';index" json:"status"` // 'wishlist', 'approved', 'booked'
	Quantity      int       `gorm:"not null;default:1" json:"quantity"`             // Number of units
	LockedPrice   float64   `gorm:"type:decimal(10,2)" json:"locked_price"`         // Price at time of adding
	Notes         string    `gorm:"type:text" json:"notes,omitempty"`               // Optional notes
	AddedBy       uuid.UUID `gorm:"type:uuid;index" json:"added_by"`                // User who added this item
	User          User      `gorm:"foreignKey:AddedBy" json:"-"`                    // Reference to user who added
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BeforeSave validates the cart item before saving
func (c *CartItem) BeforeSave(tx *gorm.DB) error {
	// Validate type
	validTypes := map[string]bool{
		"room":     true,
		"banquet":  true,
		"catering": true,
		"flight":   true,
		"hotel":    true,
	}
	if !validTypes[c.Type] {
		return gorm.ErrInvalidData
	}

	// Validate status
	validStatuses := map[string]bool{
		"wishlist": true,
		"approved": true,
		"booked":   true,
	}
	if !validStatuses[c.Status] {
		return gorm.ErrInvalidData
	}

	return nil
}

// --- Response DTOs for Hierarchical Cart Response ---

// CartResponse is the top-level response structure
type CartResponse struct {
	EventID uuid.UUID        `json:"event_id"`
	Hotels  []HotelCartGroup `json:"hotels"`
	Flights []CartItemDetail `json:"flights"`
}

// HotelCartGroup groups cart items by hotel
type HotelCartGroup struct {
	HotelDetails      interface{}      `json:"hotel_details"` // Full Hotel object
	Rooms             []CartItemDetail `json:"rooms"`
	Banquets          []CartItemDetail `json:"banquets"`
	Catering          []CartItemDetail `json:"catering"`
	HotelWishlistItem *CartItemDetail  `json:"hotel_wishlist_item,omitempty"`
}

// CartItemDetail combines cart item with referenced item details
type CartItemDetail struct {
	// Cart item fields
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Quantity    int       `json:"quantity"`
	LockedPrice float64   `json:"locked_price"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// Referenced item details (polymorphic - populated based on Type)
	RoomDetails     interface{} `json:"room_details,omitempty"`
	BanquetDetails  interface{} `json:"banquet_details,omitempty"`
	CateringDetails interface{} `json:"catering_details,omitempty"`
	FlightDetails   interface{} `json:"flight_details,omitempty"`
	HotelDetails    interface{} `json:"hotel_details,omitempty"`
}
