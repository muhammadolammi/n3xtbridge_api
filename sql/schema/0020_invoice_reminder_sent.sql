-- +goose Up 
ALTER TABLE invoices
ADD reminder_sent_at TIMESTAMP  ;

-- +goose Down
ALTER TABLE invoices
DROP COLUMN reminder_sent_at;