CREATE TABLE IF NOT EXISTS apps_cache (
    id UInt32,
    app String,
    name String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY id;
