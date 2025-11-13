-- +goose Up
-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_topics_name ON topics(name);
CREATE INDEX IF NOT EXISTS idx_subscribers_email ON subscribers(email);
CREATE INDEX IF NOT EXISTS idx_subscribers_active ON subscribers(is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_subscriber_id ON subscriptions(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_topic_id ON subscriptions(topic_id);
CREATE INDEX IF NOT EXISTS idx_contents_topic_id ON contents(topic_id);
CREATE INDEX IF NOT EXISTS idx_contents_published ON contents(is_published);
CREATE INDEX IF NOT EXISTS idx_contents_published_at ON contents(published_at);
CREATE INDEX IF NOT EXISTS idx_email_logs_subscriber_id ON email_logs(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_content_id ON email_logs(content_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);
CREATE INDEX IF NOT EXISTS idx_email_logs_sent_at ON email_logs(sent_at);

-- +goose Down
DROP INDEX IF EXISTS idx_topics_name;
DROP INDEX IF EXISTS idx_subscribers_email;
DROP INDEX IF EXISTS idx_subscribers_active;
DROP INDEX IF EXISTS idx_subscriptions_subscriber_id;
DROP INDEX IF EXISTS idx_subscriptions_topic_id;
DROP INDEX IF EXISTS idx_contents_topic_id;
DROP INDEX IF EXISTS idx_contents_published;
DROP INDEX IF EXISTS idx_contents_published_at;
DROP INDEX IF EXISTS idx_email_logs_subscriber_id;
DROP INDEX IF EXISTS idx_email_logs_content_id;
DROP INDEX IF EXISTS idx_email_logs_status;
DROP INDEX IF EXISTS idx_email_logs_sent_at;