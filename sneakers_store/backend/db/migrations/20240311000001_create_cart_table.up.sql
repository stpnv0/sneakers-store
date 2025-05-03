CREATE TABLE IF NOT EXISTS cart (
    id SERIAL PRIMARY KEY,
    user_sso_id INTEGER NOT NULL,
    sneaker_id INTEGER NOT NULL REFERENCES sneakers(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    UNIQUE (user_sso_id, sneaker_id)
);
