-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_subscribers_email ON subscribers(email);
CREATE INDEX IF NOT EXISTS idx_contents_topic_id ON contents(topic_id);
CREATE INDEX IF NOT EXISTS idx_contents_send_time ON contents(send_time);
CREATE INDEX IF NOT EXISTS idx_subscriptions_subscriber_topic ON subscriptions(subscriber_id, topic_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_subscriber_id ON email_logs(subscriber_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_subscribers_email;
DROP INDEX IF EXISTS idx_contents_topic_id;
DROP INDEX IF EXISTS idx_contents_send_time;
DROP INDEX IF EXISTS idx_subscriptions_subscriber_topic;
DROP INDEX IF EXISTS idx_email_logs_subscriber_id;
-- +goose StatementEnd
