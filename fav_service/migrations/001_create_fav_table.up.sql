CREATE TABLE IF NOT EXISTS favourites_items (
    id SERIAL PRIMARY KEY,
    user_sso_id INTEGER NOT NULL,
    sneaker_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_sneaker UNIQUE (user_sso_id, sneaker_id)
);

CREATE INDEX IF NOT EXISTS idx_favourites_user_sso_id ON favourites_items (user_sso_id);

