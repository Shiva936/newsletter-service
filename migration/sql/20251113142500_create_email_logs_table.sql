-- +goose Up
-- Create email_logs table
CREATE TABLE IF NOT EXISTS email_logs (
    id SERIAL PRIMARY KEY,
    subscriber_id INTEGER NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    content_id INTEGER NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    email_address VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    sent_at TIMESTAMP WITH TIME ZONE NULL,
    error_message TEXT NULL,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS email_logs;