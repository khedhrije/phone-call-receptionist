CREATE TABLE sms_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES inbound_calls(id),
    to_phone VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    twilio_sid VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sms_logs_call_id ON sms_logs(call_id);
