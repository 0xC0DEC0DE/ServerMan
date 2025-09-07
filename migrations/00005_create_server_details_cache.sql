CREATE TABLE IF NOT EXISTS server_details_cache (
    server_id UInt32,
    name String,
    state String,
    ip_address String,
    operating_system String,
    memory String,
    disk String,
    cpu String,
    vnc_status String,
    daily_snapshots String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY server_id;
