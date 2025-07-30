CREATE TABLE IF NOT EXISTS sneakers (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    price REAL NOT NULL,
    image_key VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_sneakers_title ON sneakers (title);