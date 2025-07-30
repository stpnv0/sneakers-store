import styles from './styles.module.scss';
import { memo, useContext, useCallback } from 'react';
import { FavoritesContext } from '../../context/FavoritesContext';
import { logger } from '../../utils/logger';

export const HeartButton = memo(({ sneakerId }) => {
  if (!sneakerId) {
    logger.error('HeartButton: отсутствует sneakerId');
    return null;
  }
  
  const { isFavorite, toggleFavorite } = useContext(FavoritesContext);
  const isLiked = isFavorite(sneakerId);

  const handleClick = useCallback(async (e) => {
    e.preventDefault(); // Предотвращаем всплытие события
    try {
      await toggleFavorite(sneakerId);
    } catch (error) {
      logger.error('Ошибка при изменении статуса избранного:', error);
    }
  }, [sneakerId, toggleFavorite]);

  return (
    <button 
      className={styles.favorite} 
      onClick={handleClick}
      aria-label={isLiked ? "Remove from favourites" : "Add to favourites"}
    >
      <img 
        width={35} 
        height={35} 
        src={isLiked ? "/img/liked.svg" : "/img/unliked.svg"} 
        alt="heart" 
      />
    </button>
  );
});