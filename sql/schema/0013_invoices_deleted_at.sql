-- +goose Up
ALTER TABLE invoices ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_invoices_deleted_at ON invoices(deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_invoices_deleted_at;
ALTER TABLE invoices DROP COLUMN deleted_at;