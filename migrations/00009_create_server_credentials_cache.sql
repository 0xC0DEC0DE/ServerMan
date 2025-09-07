CREATE TABLE IF NOT EXISTS server_credentials_cache (
    server_id UInt32,
    root_password String,
    vnc_host String,
    vnc_port UInt16,
    vnc_password String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY server_id;
