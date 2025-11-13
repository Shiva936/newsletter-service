-- +goose Up
-- Add notification tracking fields to contents table
ALTER TABLE contents 
ADD COLUMN notifications_sent BOOLEAN DEFAULT FALSE,
ADD COLUMN notifications_sent_at TIMESTAMP WITH TIME ZONE NULL;

-- Add index for efficient querying of pending notifications
CREATE INDEX idx_contents_notifications_sent ON contents(notifications_sent);
CREATE INDEX idx_contents_published_notifications ON contents(is_published, notifications_sent);

-- +goose Down
-- Remove notification tracking fields and indexes
DROP INDEX IF EXISTS idx_contents_published_notifications;
DROP INDEX IF EXISTS idx_contents_notifications_sent;
ALTER TABLE contents 
DROP COLUMN IF EXISTS notifications_sent_at,
DROP COLUMN IF EXISTS notifications_sent;