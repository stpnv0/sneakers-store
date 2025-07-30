import axios from 'axios';
import { logger } from '../utils/logger';

// Создаем новый экземпляр axios с базовым URL
const instance = axios.create({
  baseURL: 'http://localhost',
  headers: {
    'Content-Type': 'application/json'
  }
});

// Функция для маскирования токена
const maskToken = (token) => {
  if (!token) return '';
  if (token.length < 10) return '***';
  return `${token.slice(0, 4)}...${token.slice(-4)}`;
};

// Добавляем перехватчик запросов для добавления токена авторизации
instance.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token');
    if (token) {
      // Не выводим токен в консоль, чтобы не раскрывать конфиденциальную информацию
      logger.info('Отправляем авторизованный запрос', { 
        url: config.url, 
        method: config.method
      });
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    logger.error('Ошибка при настройке запроса:', error);
    return Promise.reject(error);
  }
);

// Добавляем перехватчик ответов для обработки ошибок
instance.interceptors.response.use(
  response => {
    return response;
  },
  error => {
    if (error.response) {
      logger.info('Ошибка ответа:', { 
        status: error.response.status, 
        data: error.response.data 
      });
      
      // Если сервер вернул ошибку 401 (не авторизован), можно обработать ее здесь
      if (error.response.status === 401) {
        logger.warn('Ошибка авторизации', { 
          url: error.config.url, 
          method: error.config.method 
        });
      }
    } else if (error.request) {
      logger.warn('Нет ответа от сервера');
    } else {
      logger.error('Ошибка при настройке запроса:', { message: error.message });
    }
    return Promise.reject(error);
  }
);

export default instance; 