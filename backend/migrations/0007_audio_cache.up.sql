CREATE TABLE audio_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hash VARCHAR(64) UNIQUE NOT NULL,
    voice_id VARCHAR(255) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audio_cache_hash ON audio_cache(hash);
