-- +goose Up
ALTER TABLE sneakers ALTER COLUMN price TYPE BIGINT USING price::BIGINT;

-- +goose Down
ALTER TABLE sneakers ALTER COLUMN price TYPE REAL USING price::REAL;
