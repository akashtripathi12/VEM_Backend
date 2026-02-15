-- Migration: Create cart_items table
-- Description: Unified cart/wishlist system for event planning

CREATE TABLE IF NOT EXISTS cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('room', 'banquet', 'catering', 'flight')),
    ref_id VARCHAR(255) NOT NULL,
    parent_hotel_id VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'wishlist' CHECK (status IN ('wishlist', 'approved', 'booked')),
    quantity INTEGER NOT NULL DEFAULT 1,
    locked_price DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    added_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_cart_items_event_id ON cart_items(event_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_type ON cart_items(type);
CREATE INDEX IF NOT EXISTS idx_cart_items_status ON cart_items(status);
CREATE INDEX IF NOT EXISTS idx_cart_items_parent_hotel_id ON cart_items(parent_hotel_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_added_by ON cart_items(added_by);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_cart_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_cart_items_updated_at
    BEFORE UPDATE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_cart_items_updated_at();

-- Add comments for documentation
COMMENT ON TABLE cart_items IS 'Stores cart/wishlist items for events with polymorphic references to rooms, banquets, catering, and flights';
COMMENT ON COLUMN cart_items.type IS 'Type of item: room, banquet, catering, or flight';
COMMENT ON COLUMN cart_items.ref_id IS 'ID of the referenced item in the respective table';
COMMENT ON COLUMN cart_items.parent_hotel_id IS 'Hotel code for grouping (NULL for flights)';
COMMENT ON COLUMN cart_items.status IS 'Status: wishlist (draft), approved (final cart), or booked';
COMMENT ON COLUMN cart_items.locked_price IS 'Price at the time of adding to cart';
