-- +goose Up
CREATE TABLE IF NOT EXISTS alerts (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'default',
    duration INT DEFAULT 5000,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_project_id ON alerts(project_id);

-- +goose Down
DROP TABLE IF EXISTS alerts;
