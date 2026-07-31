CREATE TABLE vault_events(
    id BIGSERIAL PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    event_timestamp TIMESTAMPTZ NOT NULL,
    event_type TEXT NOT NULL,
    amount NUMERIC,
    previous_amount NUMERIC,
    new_amount NUMERIC,
    UNIQUE(tx_hash, log_index)
);
CREATE INDEX idx_vault_events_wallet_cursor ON vault_events(wallet_address, block_number, log_index);
