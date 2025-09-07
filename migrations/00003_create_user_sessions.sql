CREATE TABLE IF NOT EXISTS user_sessions (
    id UInt64,
    email String,
    session_token String,
    created_at DateTime DEFAULT now(),
) ENGINE = MergeTree()
ORDER BY id
TTL created_at + INTERVAL 12 HOUR;