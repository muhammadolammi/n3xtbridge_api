-- +goose Up 
ALTER TABLE quotes
ADD user_id UUID NOT NULL  REFERENCES users(id) ;


-- +goose Down
ALTER TABLE quotes
DROP COLUMN user_id;
