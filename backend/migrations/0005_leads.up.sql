CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES inbound_calls(id),
    phone VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255),
    email VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leads_phone ON leads(phone);
CREATE INDEX idx_leads_status ON leads(status);
