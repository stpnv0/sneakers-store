import styles from "./styles.module.scss"
import { memo, useContext } from 'react';
import { HeartButton } from '../HeartButton/HeartButton';
import { CartContext } from '../../context/CartContext';
import { logger } from '../../utils/logger';

export const Card = memo(({ id, title, description, imgUrl, price }) => {
  const { isInCart, addToCart, decreaseQuantity, getQuantity } = useContext(CartContext);
  const inCart = isInCart(id);
  const quantity = getQuantity(id);

  const handlePlusClick = async () => {
    try {
      await addToCart(id);
    } catch (error) {
      logger.error('Ошибка при добавлении в корзину', { error });
    }
  };

  const handleMinusClick = async () => {
    try {
      await decreaseQuantity(id);
    } catch (error) {
      logger.error('Ошибка при уменьшении количества', { error });
    }
  };

  const handleImageError = (e) => {
    logger.warn('Ошибка загрузки изображения', { imgUrl });
    e.target.src = '/img/sneakers/1.jpg'; // Fallback изображение
  };

  return (
    <div className={styles.card}>
      <div className={styles.favorite}>
        <HeartButton sneakerId={id} />
      </div>
      <img 
        width={133} 
        height={112} 
        src={imgUrl} 
        alt={title}
        onError={handleImageError}
      />
      <h5>{title}</h5>
      <div className={styles.cardButton}>
        <div className={styles.price}>
          <span>Цена:</span>
          <b>{price} руб.</b>
        </div>
        <div className={styles.addButton}>
          {inCart ? (
            <>
              <button 
                className={styles.quantityBtn} 
                onClick={handleMinusClick}
                aria-label="Уменьшить количество"
              >
                –
              </button>
              <span className={styles.quantityBadge}>{quantity}</span>
              <button 
                className={styles.quantityBtn} 
                onClick={handlePlusClick}
                aria-label="Увеличить количество"
              >
                +
              </button>
            </>
          ) : (
            <img
              className={styles.btn}
              onClick={handlePlusClick}
              src="/img/plus.svg"
              alt="Добавить в корзину"
            />
          )}
        </div>
      </div>
    </div>
  );
});
