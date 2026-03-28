-- +goose Up 
ALTER TABLE quote_requests
ADD budget DECIMAL(10, 2) ;

-- +goose Down
ALTER TABLE quote_requests
DROP COLUMN budget;
