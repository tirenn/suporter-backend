-- +goose Up
CREATE TABLE IF NOT EXISTS donations (
    id BIGSERIAL PRIMARY KEY,
    streamer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL,
    sender_name VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    unique_code INT NOT NULL,
    total_amount BIGINT NOT NULL,
    message TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_donations_total_amount ON donations(total_amount);
CREATE INDEX IF NOT EXISTS idx_donations_status ON donations(status);

-- +goose Down
DROP TABLE IF EXISTS donations;
