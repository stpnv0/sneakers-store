import React, { useContext } from 'react';
import { Card } from '../components/Card/Card';
import { FavoritesContext } from '../context/FavoritesContext';
import styles from './Favourites.module.scss';

// Базовый URL для изображений. Вынесите в константы или .env
const S3_BASE_URL = 'http://localhost:9000/sneakers';

export const Favourites = () => {
  // --- ГЛАВНОЕ ИЗМЕНЕНИЕ ---
  // Мы берем ТОЛЬКО `favorites` из FavoritesContext.
  // Этот массив уже содержит ПОЛНУЮ, обогащенную информацию о товарах.
  // `ItemsContext` здесь больше не нужен!
  const { favorites, loading, error } = useContext(FavoritesContext);
  // -------------------------

  if (loading) {
    return (
      <div className={styles.favourites}>
        <h1>Мои закладки</h1>
        <div className={styles.loading}>Загрузка...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.favourites}>
        <h1>Мои закладки</h1>
        <div className={styles.error}>Ошибка: {error}</div>
      </div>
    );
  }

  return (
    <div className={styles.favourites}>
      <h1>Мои закладки</h1>
      {favorites.length > 0 ? (
        <div className={styles.sneakers}>
          {/* Мы просто итерируемся по готовому массиву `favorites` */}
          {favorites.map((item) => {
            // Формируем полный URL для изображения
            const imageUrl = item.image_key
              ? `${S3_BASE_URL}/${item.image_key}`
              : '/img/placeholder.svg';

            return (
              <Card 
                key={item.id}
                id={item.id}
                title={item.title}
                price={item.price}
                imgUrl={imageUrl} // <-- Передаем полный URL
              />
            );
          })}
        </div>
      ) : (
        <div className={styles.empty}>
          <p>У вас пока нет избранных товаров.</p>
          <p>Добавьте товары в избранное, нажав на сердечко на карточке товара.</p>
        </div>
      )}
    </div>
  );
};