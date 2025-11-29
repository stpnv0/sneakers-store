import React, { createContext, useContext, useState, useEffect } from 'react';
import axios from '../api/axios';
import { logger } from '../utils/logger';
import { ItemsContext } from './ItemsContext';

export const CartContext = createContext({
  cartItems: [],
  isLoading: false,
  totalPrice: 0,
  isInCart: () => false,
  getQuantity: () => 0,
  addToCart: () => { },
  removeFromCart: () => { },
  increaseQuantity: () => { },
  decreaseQuantity: () => { },
  getTotalPrice: () => 0,
  getTaxAmount: () => 0,
  error: null
});

export const CartProvider = ({ children }) => {
  const [cartItems, setCartItems] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const { items } = useContext(ItemsContext);

  // Загрузка корзины
  const fetchCart = async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      logger.info('Пользователь не авторизован, корзина недоступна');
      setCartItems([]);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      // Добавляем timestamp для предотвращения кэширования и слеш в конце URL
      const timestamp = new Date().getTime();
      const response = await axios.get(`/api/v1/cart/`);

      logger.info('Получены данные корзины', response.data);

      let cartArray = [];
      if (response.data && Array.isArray(response.data)) {
        cartArray = response.data;
      } else if (response.data && Array.isArray(response.data.items)) {
        cartArray = response.data.items;
      } else if (response.data && Array.isArray(response.data.cart)) {
        cartArray = response.data.cart;
      } else if (response.data && typeof response.data === 'object') {
        // Если данные - это объект, но не массив, пробуем извлечь информацию
        const possibleArrays = Object.values(response.data).filter(Array.isArray);
        if (possibleArrays.length > 0) {
          cartArray = possibleArrays[0];
        }
      }

      // Логирование без вывода всех деталей товаров
      logger.info('Загружено товаров в корзине', { count: cartArray.length });

      setCartItems(cartArray);
      logger.info('Корзина загружена', { count: cartArray.length });
    } catch (err) {
      if (err.response?.status === 401) {
        logger.warn('Требуется авторизация для доступа к корзине');
      } else {
        logger.error('Ошибка при загрузке корзины', err);
      }
      setError(err.response?.data?.error || 'Не удалось загрузить корзину');
      setCartItems([]);
    } finally {
      setIsLoading(false);
    }
  };

  // Проверяет наличие товара в корзине
  const isInCart = (sneakerId) => {
    if (!Array.isArray(cartItems)) {
      console.warn('cartItems не является массивом', cartItems);
      return false;
    }
    return cartItems.some(item => item.sneaker_id === sneakerId);
  };

  // Получает количество товара в корзине
  const getQuantity = (sneakerId) => {
    if (!Array.isArray(cartItems)) {
      return 0;
    }
    const item = cartItems.find(item => item.sneaker_id === sneakerId);
    return item ? item.quantity : 0;
  };

  // Добавляет товар в корзину
  const addToCart = async (sneakerId) => {
    const token = localStorage.getItem('token');
    if (!token) {
      alert('Пожалуйста, войдите в аккаунт, чтобы добавить товар в корзину');
      return;
    }

    try {
      setIsLoading(true);

      if (isInCart(sneakerId)) {
        // Если товар уже в корзине - увеличиваем количество
        return await increaseQuantity(sneakerId);
      }

      // Добавляем товар в корзину - обратите внимание на слеш в конце URL
      const response = await axios.post('/api/v1/cart/', { sneaker_id: sneakerId, quantity: 1 });
      console.log('Ответ при добавлении в корзину:', response.data);

      // Оптимистично обновляем UI
      setCartItems(prev => [...prev, { sneaker_id: sneakerId, quantity: 1 }]);

      logger.info('Товар добавлен в корзину', { sneakerId });
      // Обновляем данные с сервера
      await fetchCart();
    } catch (err) {
      logger.error('Ошибка при добавлении товара в корзину', err);
      console.error('Детали ошибки:', err.response || err);
      setError(err.response?.data?.error || 'Не удалось добавить товар в корзину');
    } finally {
      setIsLoading(false);
    }
  };

  // Удаляет товар из корзины
  const removeFromCart = async (sneakerId) => {
    const token = localStorage.getItem('token');
    if (!token) return;

    try {
      setIsLoading(true);

      const item = cartItems.find(item => item.sneaker_id === sneakerId);
      if (!item || !item.id) {
        console.warn('Не удалось найти ID записи для удаления из корзины', item);
        return;
      }

      logger.info('Отправляем запрос на удаление товара', { itemId: item.id });
      const response = await axios.delete(`/api/v1/cart/${item.id}`);

      // Оптимистично обновляем UI
      setCartItems(prev => prev.filter(item => item.sneaker_id !== sneakerId));

      logger.info('Товар удален из корзины', { sneakerId });
    } catch (err) {
      logger.error('Ошибка при удалении товара из корзины', err);
      setError(err.response?.data?.error || 'Не удалось удалить товар из корзины');

      // Обновляем корзину с сервера в случае ошибки
      await fetchCart();
    } finally {
      setIsLoading(false);
    }
  };

  // Увеличивает количество товара в корзине
  const increaseQuantity = async (sneakerId) => {
    const token = localStorage.getItem('token');
    if (!token) return;

    try {
      setIsLoading(true);

      const item = cartItems.find(item => item.sneaker_id === sneakerId);
      if (!item || !item.id) {
        console.warn('Не удалось найти ID записи для увеличения количества', item);
        return;
      }

      const newQuantity = item.quantity + 1;

      // Оптимистично обновляем UI
      setCartItems(prev => prev.map(cartItem =>
        cartItem.sneaker_id === sneakerId
          ? { ...cartItem, quantity: newQuantity }
          : cartItem
      ));

      console.log(`Отправляем запрос на увеличение количества: PUT /api/v1/cart/${item.id}`, {
        sneaker_id: sneakerId,
        quantity: newQuantity
      });

      const response = await axios.put(`/api/v1/cart/${item.id}`, {
        sneaker_id: sneakerId,
        quantity: newQuantity
      });

      console.log('Ответ сервера на увеличение количества:', response.data);
      logger.info('Увеличено количество товара в корзине', { sneakerId, quantity: newQuantity });
    } catch (err) {
      logger.error('Ошибка при увеличении количества товара', err);
      console.error('Детали ошибки:', err.response || err);
      setError(err.response?.data?.error || 'Не удалось увеличить количество товара');

      // Обновляем корзину с сервера в случае ошибки
      await fetchCart();
    } finally {
      setIsLoading(false);
    }
  };

  // Уменьшает количество товара в корзине
  const decreaseQuantity = async (sneakerId) => {
    const token = localStorage.getItem('token');
    if (!token) return;

    try {
      setIsLoading(true);

      const item = cartItems.find(item => item.sneaker_id === sneakerId);
      if (!item || !item.id) {
        console.warn('Не удалось найти ID записи для уменьшения количества', item);
        return;
      }

      // Если количество = 1, удаляем товар из корзины
      if (item.quantity <= 1) {
        return await removeFromCart(sneakerId);
      }

      const newQuantity = item.quantity - 1;

      // Оптимистично обновляем UI
      setCartItems(prev => prev.map(cartItem =>
        cartItem.sneaker_id === sneakerId
          ? { ...cartItem, quantity: newQuantity }
          : cartItem
      ));

      logger.info('Отправляем запрос на уменьшение количества', {
        itemId: item.id,
        newQuantity
      });

      const response = await axios.put(`/api/v1/cart/${item.id}`, {
        sneaker_id: sneakerId,
        quantity: newQuantity
      });

      logger.info('Уменьшено количество товара в корзине', { sneakerId, quantity: newQuantity });
    } catch (err) {
      logger.error('Ошибка при уменьшении количества товара', err);
      setError(err.response?.data?.error || 'Не удалось уменьшить количество товара');

      // Обновляем корзину с сервера в случае ошибки
      await fetchCart();
    } finally {
      setIsLoading(false);
    }
  };

  // Расчет общей стоимости товаров в корзине
  const getTotalPrice = () => {
    if (!Array.isArray(cartItems) || !Array.isArray(items)) {
      return 0;
    }

    return cartItems.reduce((sum, cartItem) => {
      const product = items.find(item => item.id === cartItem.sneaker_id);
      return sum + (product ? product.price * cartItem.quantity : 0);
    }, 0);
  };

  // Расчет налога (5% от общей суммы)
  const getTaxAmount = () => {
    return Math.round(getTotalPrice() * 0.05);
  };

  // Инициализация
  useEffect(() => {
    fetchCart();
  }, []);

  return (
    <CartContext.Provider value={{
      cartItems,
      isLoading,
      error,
      totalPrice: getTotalPrice(),
      isInCart,
      getQuantity,
      addToCart,
      removeFromCart,
      increaseQuantity,
      decreaseQuantity,
      getTotalPrice,
      getTaxAmount
    }}>
      {children}
    </CartContext.Provider>
  );
}; 