-- +goose Up

-- Create service_orders table for solar and security camera installation orders
CREATE TABLE IF NOT EXISTS service_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    whatsapp_phone VARCHAR(50),
    company_size VARCHAR(20) NOT NULL CHECK (company_size IN ('sole_proprietor', '1-10', '11-50', '51-200', '201-1000', '1000+')),
    referral_source VARCHAR(255) NOT NULL,
    service_type VARCHAR(20) NOT NULL, -- CHECK (service_type IN ('solar', 'security','tech')),
    appliance_details JSONB,  -- Array of {name, quantity, estimated_cost, partner_vendor}
    delivery_address TEXT NOT NULL,
    transport_fee DECIMAL(10,2) NOT NULL,
    service_fee DECIMAL(10,2) NOT NULL,
    promo_applied BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'contacted', 'scheduled', 'completed', 'cancelled')),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_service_orders_email ON service_orders(email);
CREATE INDEX IF NOT EXISTS idx_service_orders_status ON service_orders(status);
CREATE INDEX IF NOT EXISTS idx_service_orders_user_id ON service_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_service_orders_created_at ON service_orders(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_service_orders_created_at;
DROP INDEX IF EXISTS idx_service_orders_user_id;
DROP INDEX IF EXISTS idx_service_orders_status;
DROP INDEX IF EXISTS idx_service_orders_email;
DROP TABLE IF EXISTS service_orders;
