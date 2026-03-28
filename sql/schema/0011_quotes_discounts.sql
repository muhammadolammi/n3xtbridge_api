-- +goose Up 
ALTER TABLE quotes
ADD discounts JSONB NOT NULL DEFAULT '[]' ;

-- +goose Down
ALTER TABLE quotes
DROP COLUMN discounts;
