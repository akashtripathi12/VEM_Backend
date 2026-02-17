-- Migration: Create negotiation_sessions and negotiation_rounds tables
-- Description: Support for Agent-Hotel Manager negotiation workflow

CREATE TYPE negotiation_status AS ENUM ('draft', 'waiting_for_agent', 'waiting_for_hotel', 'locked');
CREATE TYPE negotiation_modifier AS ENUM ('agent', 'hotel');
CREATE TYPE negotiation_reason AS ENUM ('budget_constraint', 'volume_discount', 'competitor_offer', 'other');

CREATE TABLE IF NOT EXISTS negotiation_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    status negotiation_status NOT NULL DEFAULT 'draft',
    share_token UUID UNIQUE DEFAULT gen_random_uuid(),
    current_round INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_negotiation_sessions_event_id ON negotiation_sessions(event_id);
CREATE INDEX IF NOT EXISTS idx_negotiation_sessions_share_token ON negotiation_sessions(share_token);

CREATE TABLE IF NOT EXISTS negotiation_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES negotiation_sessions(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    modified_by negotiation_modifier NOT NULL,
    proposal_snapshot JSONB NOT NULL, -- Stores array of items with prices
    remarks TEXT,
    reason_code negotiation_reason,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_negotiation_rounds_session_id ON negotiation_rounds(session_id);

-- Trigger to update updated_at for negotiation_sessions
CREATE OR REPLACE FUNCTION update_negotiation_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_negotiation_sessions_updated_at
    BEFORE UPDATE ON negotiation_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_negotiation_sessions_updated_at();

-- Comments
COMMENT ON TABLE negotiation_sessions IS 'Tracks a negotiation process between Agent and Hotel for an Event';
COMMENT ON COLUMN negotiation_sessions.share_token IS 'Unique token for Hotel Manager access without login';
COMMENT ON TABLE negotiation_rounds IS 'Immutable record of each turn in the negotiation';
COMMENT ON COLUMN negotiation_rounds.proposal_snapshot IS 'JSON blob snapshot of all line items and their proposed prices';
