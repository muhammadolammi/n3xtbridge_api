-- +goose Up

ALTER TABLE service_categories
ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE service_categories
ADD COLUMN icon TEXT NOT NULL DEFAULT 'Briefcase';

-- +goose Down
ALTER TABLE service_categories
drop  COLUMN status;
ALTER TABLE service_categories
drop  COLUMN icon;
    