CREATE TABLE IF NOT EXISTS users (
    id UInt64,
    email String,
    groups Array(String),
) ENGINE = ReplacingMergeTree()
PRIMARY KEY (email, groups);
