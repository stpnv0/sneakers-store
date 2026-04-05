-- +goose Up
INSERT INTO sneakers (title, price, image_key) VALUES
('Кроссовки Nike Blazer 77 Suede', 799900, 'products/1.jpg'),
('Кроссовки Nike Air Max 270', 699900, 'products/2.jpg'),
('Кроссовки Nike Blazer 77 White', 899900, 'products/3.jpg'),
('Puma Future Ride', 599900, 'products/4.jpg'),
('Женские Кроссовки Demix CURRY 8', 849900, 'products/5.jpg'),
('Кроссовки Nike Kyrie 7', 1149900, 'products/6.jpg'),
('Кроссовки Air Jordan 11', 1449900, 'products/7.jpg'),
('Кроссовки Nike Lebron 18', 1949900, 'products/8.jpg'),
('Кроссовки Nike Blazer Green', 999900, 'products/1.jpg');

-- +goose Down
DELETE FROM sneakers WHERE title IN (
    'Кроссовки Nike Blazer 77 Suede',
    'Кроссовки Nike Air Max 270',
    'Кроссовки Nike Blazer 77 White',
    'Puma Future Ride',
    'Женские Кроссовки Demix CURRY 8',
    'Кроссовки Nike Kyrie 7',
    'Кроссовки Air Jordan 11',
    'Кроссовки Nike Lebron 18',
    'Кроссовки Nike Blazer Green'
);
