-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; 

CREATE TABLE service_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    description TEXT UNIQUE NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   
);
ALTER TABLE services
ADD COLUMN category_id UUID REFERENCES service_categories(id)
ON DELETE SET NULL;
ALTER TABLE services
drop  COLUMN category;

-- +goose Down
DROP TABLE service_categories;
ALTER TABLE services
drop  COLUMN category_id;
