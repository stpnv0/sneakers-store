import React, { useContext } from 'react';
import { CartContext } from '../context/CartContext';
import { ItemsContext } from '../context/ItemsContext';

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
  
  const { items: itemsData } = useContext(ItemsContext);

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
    // Проверяем, что клик был именно по overlay, а не по его дочерним элементам
    if (e.target.className === 'overlay') {
      onClose();
    }
  };

  return (
    <div className="overlay" onClick={handleOverlayClick}>
      <div className="drawer">
        <h2>
          Корзина <img onClick={onClose} className="btnRemove" src="/img/btnRemove.svg" alt="Закрыть" />
        </h2>

        {cartItems && cartItems.length > 0 ? (
          <>
            <div className="items">
              {cartItems.map((item) => {
                const sneaker = itemsData.find(s => s.id === item.sneaker_id);
                if (!sneaker) {
                  console.warn(`Товар с ID ${item.sneaker_id} не найден в списке товаров`);
                  return null;
                }
                
                return (
                  <div key={item.id} className="cartItem">
                    <img 
                      className="cartItemImg" 
                      src={sneaker.imageUrl} 
                      alt={sneaker.title} 
                      onError={(e) => {
                        e.target.onerror = null;
                        e.target.src = "/img/sneakersPlaceholder.jpg";
                      }}
                    />
                    <div className="description">
                      <p>{sneaker.title}</p>
                      <b>{sneaker.price} руб.</b>
                      <div className="cartItemQuantity">
                        <div className="quantityControl">
                          <button 
                            onClick={() => handleCartAction('decrease', item.sneaker_id)}
                            disabled={isLoading}
                          >
                            -
                          </button>
                          <span>{item.quantity}</span>
                          <button 
                            onClick={() => handleCartAction('increase', item.sneaker_id)}
                            disabled={isLoading}
                          >
                            +
                          </button>
                        </div>
                      </div>
                    </div>
                    <button 
                      className="removeBtn" 
                      onClick={() => handleCartAction('remove', item.sneaker_id)}
                      disabled={isLoading}
                    >
                      <img src="/img/btnRemove.svg" alt="Удалить" />
                    </button>
                  </div>
                );
              })}
            </div>
            <div className="cartTotalBlock">
              <ul>
                <li>
                  <span>Итого: </span>
                  <div></div>
                  <b>{getTotalPrice()} руб.</b>
                </li>
                <li>
                  <span>Налог 5%: </span>
                  <div></div>
                  <b>{getTaxAmount()} руб.</b>
                </li>
              </ul>
              <button className="greenBtn" disabled={isLoading}>
                Оформить заказ <img src="/img/arrow.svg" alt="Стрелка" />
              </button>
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