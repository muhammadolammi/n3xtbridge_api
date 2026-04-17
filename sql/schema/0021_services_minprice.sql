-- +goose Up 
ALTER TABLE services
-- this is not useful anymore
DROP COLUMN IF EXISTS icon,
ADD min_price TEXT  NOT NULL DEFAULT '' ;

-- +goose Down
ALTER TABLE services
DROP COLUMN  IF EXISTS min_price;