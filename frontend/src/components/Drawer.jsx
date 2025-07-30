import React, { useContext } from 'react';
import { CartContext } from '../context/CartContext';
import { ItemsContext } from '../context/ItemsContext';

// Базовый URL для изображений.
const S3_BASE_URL = 'http://localhost:9000/sneakers';

export const Drawer = ({ onClose }) => {
  const { 
    cartItems, 
    removeFromCart, 
    decreaseQuantity, 
    increaseQuantity, 
    getTotalPrice, 
    getTaxAmount,
    isLoading
  } = useContext(CartContext);
  
  const { items: allItemsData } = useContext(ItemsContext);

  const handleCartAction = async (action, sneakerId) => {
    console.log(`Выполняем действие ${action} для товара с ID=${sneakerId}`);
    
    try {
      switch (action) {
        case 'increase':
          console.log(`Вызываем increaseQuantity(${sneakerId})`);
          await increaseQuantity(sneakerId);
          break;
        case 'remove':
          console.log(`Вызываем removeFromCart(${sneakerId})`);
          await removeFromCart(sneakerId);
          break;
        case 'decrease':
          console.log(`Вызываем decreaseQuantity(${sneakerId})`);
          await decreaseQuantity(sneakerId);
          break;
        default:
          console.log('Неизвестное действие');
          return;
      }
      console.log(`Действие ${action} выполнено успешно`);
    } catch (error) {
      console.error(`Ошибка при выполнении действия ${action}:`, error);
    }
  };

  // Обработчик клика на overlay
  const handleOverlayClick = (e) => {
    if (e.target.className === 'overlay') {
      onClose();
    }
  };

  return (
    <div className="overlay" onClick={handleOverlayClick}>
      <div className="drawer">
        <h2>Корзина <img onClick={onClose} className="btnRemove" src="/img/btnRemove.svg" alt="Закрыть" /></h2>
        {cartItems && cartItems.length > 0 ? (
          <>
            <div className="items">
              {cartItems.map((cartItem) => {
                const productDetails = allItemsData.find(p => p.id === cartItem.sneaker_id);

                if (!productDetails) {
                  return null;
                }

                const imageUrl = productDetails.image_key
                  ? `${S3_BASE_URL}/${productDetails.image_key}`
                  : "/img/sneakersPlaceholder.jpg";
                
                return (
                  <div key={cartItem.id} className="cartItem">
                    <img 
                      className="cartItemImg" 
                      src={imageUrl}
                      alt={productDetails.title} 
                    />
                    <div className="description">
                      <p>{productDetails.title}</p> 
                      <b>{productDetails.price} руб.</b>
                      <div className="cartItemQuantity">
                        <div className="quantityControl">
                          <button onClick={() => decreaseQuantity(cartItem.sneaker_id)}>-</button>
                          <span>{cartItem.quantity}</span>
                          <button onClick={() => increaseQuantity(cartItem.sneaker_id)}>+</button>
                        </div>
                      </div>
                    </div>
                    <button className="removeBtn" onClick={() => removeFromCart(cartItem.sneaker_id)}>
                      <img src="/img/btnRemove.svg" alt="Удалить" />
                    </button>
                  </div>
                );
              })}
            </div>
            <div className="cartTotalBlock">
              <ul>
                <li><span>Итого:</span><div></div><b>{getTotalPrice()} руб.</b></li>
                <li><span>Налог 5%:</span><div></div><b>{getTaxAmount()} руб.</b></li>
              </ul>
              <button className="greenBtn" disabled={isLoading}>Оформить заказ</button>
            </div>
          </>
        ) : (
          <div className="emptyCart">
            <img src="/img/empty-cart.jpg" alt="Пустая корзина" width="120" height="120" />
            <h2>Корзина пустая</h2>
            <p>Добавьте хотя бы одну пару кроссовок, чтобы сделать заказ.</p>
            <button onClick={onClose} className="greenBtn">
              <img src="/img/arrow.svg" alt="Стрелка" className="rotateArrow" />
              Вернуться назад
            </button>
          </div>
        )}
      </div>
    </div>
  );
};