-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    event_type VARCHAR(50) DEFAULT 'donation',
    html_template TEXT NOT NULL DEFAULT '',
    css_style TEXT DEFAULT '',
    fields TEXT NOT NULL DEFAULT '["name","amount","message"]',
    duration INT DEFAULT 7000,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_uuid ON projects(uuid);
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

-- +goose Down
DROP TABLE IF EXISTS projects;
