import React, { createContext, useState, useEffect, useCallback } from 'react';
import axios from '../api/axios';
import { logger } from '../utils/logger';

export const FavoritesContext = createContext({
  favoriteIds: [],
  favorites: [],
  isFavorite: () => false,
  toggleFavorite: () => {},
  loading: false,
  error: null,
  favourites: []
});

export const FavoritesProvider = ({ children }) => {
  const [favoriteIds, setFavoriteIds] = useState([]);
  const [favorites, setFavorites] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Шаг 1: получить список ID из сервиса избранного
  const fetchFavoriteIds = useCallback(async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      logger.info('User not authenticated, favorites unavailable');
      setFavoriteIds([]);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Вызываем API избранного с добавлением слеша в конце URL
      const response = await axios.get('/api/v1/favourites/');
      
      // Логируем полученные данные для отладки
      logger.info('Raw favorites response:', response.data);
      
      // Обрабатываем данные в зависимости от формата
      let ids = [];
      const data = response.data;
      
      if (Array.isArray(data) && data.length > 0) {
        // Если это массив чисел
        if (typeof data[0] === 'number') {
          ids = data;
        } 
        // Если это массив объектов
        else if (typeof data[0] === 'object') {
          ids = data.map(item => item.sneaker_id || item.id);
        }
      }
      
      logger.info('Processed favorite IDs:', ids);
      setFavoriteIds(ids);
      
    } catch (err) {
      logger.error('Error fetching favorite IDs', err.response || err);
      setError(err.response?.status === 404 ? 'Favorites service not found' : err.response?.data?.error);
      setFavoriteIds([]);
    } finally {
      setLoading(false);
    }
  }, []);

  // Шаг 2: получить детали товаров
  const fetchFavoriteDetails = useCallback(async (ids) => {
    if (!ids || ids.length === 0) {
      setFavorites([]);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Формируем строку запроса с ID товаров
      const idsString = ids.join(',');
      logger.info('Fetching details for IDs:', idsString);
      
      // Используем GET запрос с параметрами к новому эндпоинту
      const response = await axios.get(`/api/v1/products/batch?ids=${idsString}`);
      
      // Логируем ответ для отладки
      logger.info('Items response:', response.data);
      
      // Обрабатываем данные
      let products = [];
      if (Array.isArray(response.data)) {
        products = response.data;
      } else if (response.data && Array.isArray(response.data.sneakers)) {
        products = response.data.sneakers;
      }
      
      setFavorites(products);
      logger.info('Processed favorite items:', products.length);
      
    } catch (err) {
      logger.error('Error fetching favorite details', err.response || err);
      setError(err.response?.data?.error || 'Failed to load product details');
      setFavorites([]);
    } finally {
      setLoading(false);
    }
  }, []);

  // Инициализация и реакции на изменение ID
  useEffect(() => { 
    fetchFavoriteIds(); 
  }, [fetchFavoriteIds]);
  
  useEffect(() => { 
    if (favoriteIds.length > 0) {
      fetchFavoriteDetails(favoriteIds); 
    } else {
      setFavorites([]);
    }
  }, [favoriteIds, fetchFavoriteDetails]);

  const isFavorite = useCallback((id) => {
    return favoriteIds.includes(Number(id));
  }, [favoriteIds]);

  const toggleFavorite = useCallback(async (itemId) => {
    const token = localStorage.getItem('token');
    if (!token) {
      alert('Please log in to modify favorites');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Преобразуем itemId в число для корректного сравнения
      const numericItemId = Number(itemId);
      
      if (isFavorite(numericItemId)) {
        // Удаляем товар из избранного
        logger.info('Removing from favorites:', numericItemId);
        await axios.delete(`/api/v1/favourites/${numericItemId}/`);
      } else {
        // Добавляем товар в избранное
        logger.info('Adding to favorites:', numericItemId);
        await axios.post('/api/v1/favourites/', { sneaker_id: numericItemId });
      }

      // После изменения обновляем список ID
      await fetchFavoriteIds();
      
    } catch (err) {
      logger.error('Error toggling favorite', err.response || err);
      setError(err.response?.data?.error || 'Failed to update favorite');
    } finally {
      setLoading(false);
    }
  }, [isFavorite, fetchFavoriteIds]);

  return (
    <FavoritesContext.Provider value={{
      favoriteIds,
      favorites,
      favourites: favoriteIds,
      isFavorite,
      toggleFavorite,
      loading,
      error
    }}>
      {children}
    </FavoritesContext.Provider>
  );
};