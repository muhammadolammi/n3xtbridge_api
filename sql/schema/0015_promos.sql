-- +goose Up
CREATE TABLE IF NOT EXISTS promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE, -- e.g., 'FREE-INSTALL'
    name TEXT NOT NULL,               -- e.g., 'Vision for Zero Promo'
    description TEXT,
    breakdown JSONB NOT NULL DEFAULT '[]', -- [{name: "Camera", amount: 500, "discription":"short brief discription", "type":"fixed/percentage/itemname", item_name:"item name to remove from items e.g service fee"}]
    is_active BOOLEAN DEFAULT true,
    starts_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Link promos to services
ALTER TABLE services ADD COLUMN IF NOT EXISTS active_promo_ids JSONB NOT NULL DEFAULT '[]'; -- list of promos ids

-- +goose Down
ALTER TABLE services DROP COLUMN IF EXISTS active_promo_ids;
DROP TABLE IF EXISTS promotions;