CREATE TABLE system_settings (
    id INT PRIMARY KEY DEFAULT 1,
    default_llm_provider VARCHAR(100) NOT NULL DEFAULT 'gemini',
    default_voice_id VARCHAR(255) NOT NULL DEFAULT 'rachel',
    top_k INT NOT NULL DEFAULT 5,
    max_call_duration_secs INT NOT NULL DEFAULT 300,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_settings DEFAULT VALUES;
