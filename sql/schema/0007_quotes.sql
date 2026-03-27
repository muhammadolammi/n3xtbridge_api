-- +goose Up
CREATE TYPE quote_status AS ENUM ('draft', 'sent', 'accepted', 'declined', 'expired', 'in-review');
CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_request_id UUID NOT NULL UNIQUE REFERENCES quote_requests(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    breakdown JSONB NOT NULL DEFAULT '[]', -- [{item: "Camera", cost: 500}]
    notes TEXT NOT NULL DEFAULT '',
    status quote_status NOT NULL DEFAULT 'draft',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE quotes;
DROP TYPE quote_status;
