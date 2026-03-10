CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES inbound_calls(id),
    caller_phone VARCHAR(50) NOT NULL,
    caller_name VARCHAR(255) NOT NULL,
    caller_email VARCHAR(255) NOT NULL,
    service_type VARCHAR(255) NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    duration_mins INT NOT NULL DEFAULT 60,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    google_event_id VARCHAR(255),
    sms_sent_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appointments_call_id ON appointments(call_id);
CREATE INDEX idx_appointments_caller_phone ON appointments(caller_phone);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_scheduled_at ON appointments(scheduled_at);
