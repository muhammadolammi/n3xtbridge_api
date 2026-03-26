-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; 

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    description TEXT UNIQUE NOT NULL,
    category TEXT  NOT NULL,
    is_active BOOLEAN  NOT NULL DEFAULT true,
    is_featured BOOLEAN  NOT NULL,
    icon TEXT  NOT NULL,
    image TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   
);

-- +goose Down
DROP TABLE services;
