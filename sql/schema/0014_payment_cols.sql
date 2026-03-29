-- +goose Up

-- 2. Add missing columns to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS amount DECIMAL(12, 2) DEFAULT 0.00;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'NGN';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS external_id TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS metadata JSONB;

-- 3. Cleanup existing data (Optional: set amount to not null after adding it)
ALTER TABLE payments ALTER COLUMN amount SET NOT NULL;

-- +goose Down
ALTER TABLE payments DROP COLUMN IF EXISTS metadata;
ALTER TABLE payments DROP COLUMN IF EXISTS external_id;
ALTER TABLE payments DROP COLUMN IF EXISTS currency;
ALTER TABLE payments DROP COLUMN IF EXISTS amount;
-- Note: We typically don't drop types in partial rollbacks unless necessary