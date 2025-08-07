-- Создание таблицы корзин
CREATE TABLE IF NOT EXISTS carts (
    user_sso_id INTEGER PRIMARY KEY,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание таблицы элементов корзины
CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL,
    user_sso_id INTEGER NOT NULL,
    sneaker_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES carts(user_sso_id) ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_sneaker_id ON cart_items(sneaker_id); 