CREATE TABLE conversions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    original_url VARCHAR(500),
    svg_url VARCHAR(500),
    thumbnail_url VARCHAR(500),
    file_size_in BIGINT,
    file_size_out BIGINT,
    path_count INT,
    color_count INT,
    format_in VARCHAR(10),
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_conversions_user_status ON conversions(user_id, status);
CREATE INDEX idx_conversions_created ON conversions(created_at DESC);
