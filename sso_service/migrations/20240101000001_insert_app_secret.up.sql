-- Insert default app entry
-- The secret will be synced from APP_SECRET environment variable on service startup
INSERT INTO apps (id, name, secret) 
VALUES (1, 'sneakers', 'placeholder')
ON CONFLICT (id) DO NOTHING;
