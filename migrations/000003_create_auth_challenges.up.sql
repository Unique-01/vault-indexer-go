CREATE TABLE auth_challenges(
    wallet_address TEXT PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);