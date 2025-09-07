CREATE TABLE IF NOT EXISTS os_types_cache (
    id UInt32,
    name String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY id;
