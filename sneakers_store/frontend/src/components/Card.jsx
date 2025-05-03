import React, { useContext, useState, useEffect } from 'react';
import styles from './styles.module.scss';
import { CartContext } from '../context/CartContext';
import { FavoritesContext } from '../context/FavoritesContext';

export function Card({ id, title, price, imageUrl }) {
  const { 
    isInCart, 
    addToCart, 
    getCartItemQuantity,
    cartItems 
  } = useContext(CartContext);
  
  const { isFavorite, toggleFavorite } = useContext(FavoritesContext);
  const [isHovered, setIsHovered] = useState(false);
  const [loading, setLoading] = useState(false);
  
  // Отладочный вывод при изменении корзины
  useEffect(() => {
    if (isInCart(id)) {
      console.log(`Товар ${id} в корзине, количество:`, getCartItemQuantity(id));
    } else {
      console.log(`Товар ${id} не в корзине`);
    }
  }, [cartItems, id, isInCart, getCartItemQuantity]);
  
  const handlePlusClick = async () => {
    try {
      setLoading(true);
      console.log(`Нажатие на кнопку "+" для товара с ID=${id}`);
      
      const hasToken = localStorage.getItem('token');
      if (!hasToken) {
        console.log('Токен отсутствует!');
        alert('Пожалуйста, войдите в аккаунт для добавления товаров в корзину');
        return;
      }
      
      const result = await addToCart(id);
      console.log(`Результат добавления товара ${id} в корзину:`, result);
    } catch (error) {
      console.error('Ошибка при добавлении в корзину:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleFavoriteClick = () => {
    const hasToken = localStorage.getItem('token');
    if (!hasToken) {
      console.log('Токен отсутствует!');
      alert('Пожалуйста, войдите в аккаунт для добавления товаров в избранное');
      return;
    }
    
    toggleFavorite(id);
  };

  return (
    <div className={styles.card}>
      <div className={styles.favorite}>
        <img
          onClick={handleFavoriteClick}
          src={isFavorite(id) ? '/img/liked.svg' : '/img/unliked.svg'}
          alt="Like"
        />
      </div>
      <img
        width="100%"
        height={135}
        src={imageUrl}
        alt={title}
        onError={(e) => {
          e.target.onerror = null;
          e.target.src = "/img/sneakersPlaceholder.jpg";
        }}
      />
      <h5>{title}</h5>
      <div className={styles.cardBottom}>
        <div className={styles.cardPrice}>
          <span>Цена:</span>
          <b>{price} руб.</b>
        </div>
        <div
          className={`${styles.addButton} ${isInCart(id) ? styles.active : ''}`}
          onClick={handlePlusClick}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          {loading ? (
            <span>...</span>
          ) : isInCart(id) ? (
            <>
              <img src="/img/btnChecked.svg" alt="Added" />
              {getCartItemQuantity(id) > 1 && (
                <div className={styles.quantity}>{getCartItemQuantity(id)}</div>
              )}
            </>
          ) : (
            <img 
              src={isHovered ? '/img/btnPlusHover.svg' : '/img/btnPlus.svg'} 
              alt="Add" 
            />
          )}
        </div>
      </div>
    </div>
  );
} 