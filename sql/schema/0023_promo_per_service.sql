-- +goose Up

-- DROP PROMO IDS in services
ALTER TABLE services DROP COLUMN IF EXISTS active_promo_ids;
-- link promotion to a service
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS service_id UUID  REFERENCES services(id); 


-- Link promos to services

-- +goose Down
-- reverse to old state 
ALTER TABLE services ADD COLUMN IF NOT EXISTS active_promo_ids TEXT[] NOT NULL DEFAULT '{}'; -- list of promos ids
ALTER TABLE promotions DROP COLUMN IF EXISTS service_id;
