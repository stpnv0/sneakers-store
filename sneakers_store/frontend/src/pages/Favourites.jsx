import { useState, useEffect, useContext } from 'react';
import { Card } from '../components/Card/Card';
import { FavoritesContext } from '../context/FavoritesContext';
import { ItemsContext } from '../context/ItemsContext';
import styles from './Favourites.module.scss';

export const Favourites = () => {
  const { favourites, loading: favouritesLoading, error: favouritesError } = useContext(FavoritesContext);
  const { items, loading: itemsLoading } = useContext(ItemsContext);
  const [favoriteItems, setFavoriteItems] = useState([]);

  useEffect(() => {
    if (!favourites || !items) {
      setFavoriteItems([]);
      return;
    }

    try {
      console.log('Избранное:', favourites);
      console.log('Все товары:', items);
      
      // Проверяем формат данных избранного
      let favoriteIds;
      if (Array.isArray(favourites) && favourites.length > 0) {
        if (typeof favourites[0] === 'number') {
          // Если это просто массив ID
          favoriteIds = new Set(favourites);
          console.log('Избранное - массив ID:', favoriteIds);
        } else if (typeof favourites[0] === 'object' && favourites[0].sneaker_id) {
          // Если это массив объектов с полем sneaker_id
          favoriteIds = new Set(favourites.map(fav => fav.sneaker_id));
          console.log('Избранное - массив объектов:', favoriteIds);
        } else {
          console.error('Неизвестный формат данных избранного:', favourites);
          favoriteIds = new Set();
        }
      } else {
        favoriteIds = new Set();
      }
      
      // Фильтруем все товары, оставляя только те, которые в избранном
      const favoriteItems = items.filter(item => favoriteIds.has(item.id));
      console.log('Отфильтрованные избранные товары:', favoriteItems);
      
      setFavoriteItems(favoriteItems);
    } catch (err) {
      console.error('Ошибка при обработке избранных товаров:', err);
    }
  }, [favourites, items]);

  const loading = favouritesLoading || itemsLoading;
  const error = favouritesError;

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
      {favoriteItems.length > 0 ? (
        <div className={styles.sneakers}>
          {favoriteItems.map((item) => (
            <Card 
              key={item.id}
              id={item.id}
              title={item.title}
              description={item.description}
              price={item.price}
              imgUrl={item.imageUrl}
            />
          ))}
        </div>
      ) : (
        <div className={styles.empty}>
          <p>У вас пока нет избранных товаров</p>
          <p>Добавьте товары в избранное, нажав на сердечко на карточке товара</p>
        </div>
      )}
    </div>
  );
}; 