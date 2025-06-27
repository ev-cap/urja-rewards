-- users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- points_events table (append-only ledger)
CREATE TABLE points_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, -- CHARGE_KWH / REFERRAL / RATING / MANUAL_ADJUST
    ref_id TEXT, -- session-id, friend-id etc.
    points INT NOT NULL,
    meta JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- rules table for dynamic earn logic
CREATE TABLE rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- segments table for user targeting
CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    criteria JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- rewards_catalog table
CREATE TABLE rewards_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    cost INT NOT NULL,
    segment JSONB, -- user attributes or feature-flag keys
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- redemptions table
CREATE TABLE redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_id UUID NOT NULL REFERENCES rewards_catalog(id) ON DELETE CASCADE,
    points_spent INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING', -- PENDING / FULFILLED / EXPIRED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for better query performance
CREATE INDEX idx_points_events_user_id ON points_events(user_id);
CREATE INDEX idx_points_events_occurred_at ON points_events(occurred_at);
CREATE INDEX idx_points_events_event_type ON points_events(event_type);
CREATE INDEX idx_rules_active ON rules(active);
CREATE INDEX idx_rules_name ON rules(name);
CREATE INDEX idx_segments_active ON segments(active);
CREATE INDEX idx_segments_name ON segments(name);
CREATE INDEX idx_rewards_catalog_active ON rewards_catalog(active);
CREATE INDEX idx_redemptions_user_id ON redemptions(user_id);
CREATE INDEX idx_redemptions_status ON redemptions(status);
CREATE INDEX idx_redemptions_created_at ON redemptions(created_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at on redemptions
CREATE TRIGGER update_redemptions_updated_at 
    BEFORE UPDATE ON redemptions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger to automatically update updated_at on rules
CREATE TRIGGER update_rules_updated_at 
    BEFORE UPDATE ON rules 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 