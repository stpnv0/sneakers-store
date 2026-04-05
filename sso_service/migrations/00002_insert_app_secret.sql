-- +goose Up
INSERT INTO apps (id, name, secret)
VALUES (1, 'sneakers', gen_random_uuid()::text)
ON CONFLICT (id) DO UPDATE SET secret = EXCLUDED.secret WHERE apps.secret = 'placeholder';

-- +goose Down
DELETE FROM apps WHERE id = 1;
