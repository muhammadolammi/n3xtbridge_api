-- +goose Up 

-- Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50),
    total DECIMAL(10, 2) NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    items JSONB NOT NULL,
    discounts JSONB NOT NULL,
    created_at TIMESTAMP  WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create items table
-- CREATE TABLE IF NOT EXISTS items (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     invoice_id UUID REFERENCES invoices(id) ON DELETE CASCADE,
--     name VARCHAR(255) NOT NULL,
--     quantity INTEGER NOT NULL CHECK (quantity > 0),
--     price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );



-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX IF NOT EXISTS idx_invoices_customer_email ON invoices(customer_email);

-- +goose Down
DROP INDEX IF EXISTS idx_invoices_customer_email;
DROP INDEX IF EXISTS idx_invoices_invoice_number;
DROP TABLE IF EXISTS invoices;
