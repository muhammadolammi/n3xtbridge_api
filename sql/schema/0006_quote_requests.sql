-- +goose Up
CREATE TYPE quote_request_status AS ENUM ('pending', 'reviewing', 'quoted', 'rejected');

CREATE TABLE quote_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id),
    -- we should have a service name as refrence in case we delete service
    service_name TEXT NOT NULL ,
    description TEXT NOT NULL,
    attachments TEXT[] DEFAULT '{}', -- Array of URLs
    status quote_request_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);



-- +goose Down
DROP TABLE quote_requests;
DROP TYPE quote_request_status;