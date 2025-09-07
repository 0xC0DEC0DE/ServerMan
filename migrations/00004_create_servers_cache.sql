CREATE TABLE IF NOT EXISTS servers_cache (
    id UInt32,
    domain String,
    reg_date String,
    billing_cycle String,
    next_due_date String,
    domain_status String,
    last_updated DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY id;
