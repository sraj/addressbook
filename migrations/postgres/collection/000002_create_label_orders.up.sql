CREATE TABLE IF NOT EXISTS label_orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    collection_id INTEGER NOT NULL DEFAULT 0,
    contact_count INTEGER NOT NULL DEFAULT 0,
    sheet_count INTEGER NOT NULL DEFAULT 0,
    amount_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'usd',
    status TEXT NOT NULL DEFAULT 'pending',
    stripe_session_id TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);