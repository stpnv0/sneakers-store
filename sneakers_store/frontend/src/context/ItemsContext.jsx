import React, { createContext, useState, useEffect, useCallback } from 'react';
import axios from '../api/axios';
import { logger } from '../utils/logger';

export const ItemsContext = createContext({
  items: [],
  loading: false,
  error: null,
  getItemById: () => null
});

export const ItemsProvider = ({ children }) => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchItems = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await axios.get('/api/v1/items');
      console.log('Полученные данные товаров:', response.data);
      
      // Проверяем формат данных
      let itemsArray = [];
      if (Array.isArray(response.data)) {
        itemsArray = response.data;
      } else if (response.data && Array.isArray(response.data.items)) {
        itemsArray = response.data.items;
      } else if (response.data && typeof response.data === 'object') {
        const possibleArrays = Object.values(response.data).filter(Array.isArray);
        if (possibleArrays.length > 0) {
          itemsArray = possibleArrays[0];
        }
      }
      
      setItems(itemsArray);
      console.log("Загружено товаров:", itemsArray.length);
      logger.info('Товары загружены', { count: itemsArray.length });
    } catch (err) {
      console.error('Ошибка при загрузке товаров:', err.response || err);
      setError('Не удалось загрузить товары');
    } finally {
      setLoading(false);
    }
  };

  const getItemById = useCallback((id) => {
    if (!id) return null;
    // Преобразуем id в число, если передана строка
    const numericId = Number(id);
    if (isNaN(numericId)) {
      console.warn('Некорректный ID товара:', id);
      return null;
    }
    
    const item = items.find(item => item.id === numericId);
    if (!item) {
      console.warn(`Товар с ID ${numericId} не найден`);
    }
    return item || null;
  }, [items]);

  useEffect(() => {
    fetchItems();
  }, []);

  return (
    <ItemsContext.Provider value={{ items, loading, error, getItemById }}>
      {children}
    </ItemsContext.Provider>
  );
}; 