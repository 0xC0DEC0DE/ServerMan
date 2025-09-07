CREATE TABLE IF NOT EXISTS server_snapshots_cache (
    server_id UInt32,
    snapshot_id String,
    name String,
    created_at DateTime,
    size_gb Float64,
    status String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY (server_id, snapshot_id);
