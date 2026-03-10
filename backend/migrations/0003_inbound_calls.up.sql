CREATE TABLE inbound_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    twilio_call_sid VARCHAR(255) UNIQUE NOT NULL,
    caller_phone VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ringing',
    transcript JSONB NOT NULL DEFAULT '[]',
    rag_queries JSONB NOT NULL DEFAULT '[]',
    duration_seconds INT NOT NULL DEFAULT 0,
    twilio_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
    llm_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
    total_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_inbound_calls_twilio_sid ON inbound_calls(twilio_call_sid);
CREATE INDEX idx_inbound_calls_caller_phone ON inbound_calls(caller_phone);
CREATE INDEX idx_inbound_calls_status ON inbound_calls(status);
