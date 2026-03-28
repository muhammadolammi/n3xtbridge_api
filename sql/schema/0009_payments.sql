-- +goose Up 
CREATE TYPE payment_status AS ENUM ('pending', 'processing', 'success', 'failed', 'reversed');
CREATE TYPE provider_type AS ENUM ('paystack', 'manual_transfer');

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT,
    amount DECIMAL(12, 2) NOT NULL, -- The exact amount charged at the time
    currency VARCHAR(3) DEFAULT 'NGN',
    provider provider_type NOT NULL DEFAULT 'paystack',
    status payment_status NOT NULL DEFAULT 'pending',
    reference TEXT NOT NULL UNIQUE, -- Paystack Reference
    external_id TEXT, -- Paystack ID (transaction ID)
    metadata JSONB, -- Store raw provider response for debugging
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TYPE payment_status;
DROP TYPE provider_type;