import { Link } from "react-router-dom";
import { useContext, useState, useEffect } from "react";
import { CartContext } from "../context/CartContext";

export const Header = (props) => {
    const { getTotalPrice, cartItems } = useContext(CartContext);
    const [totalPrice, setTotalPrice] = useState(0);
    const [priceChanged, setPriceChanged] = useState(false);

    useEffect(() => {
        const newPrice = getTotalPrice();
        if (totalPrice !== newPrice) {
            setTotalPrice(newPrice);
            setPriceChanged(true);

            // Сбрасываем анимацию через 300 мс
            const timer = setTimeout(() => {
                setPriceChanged(false);
            }, 300);

            return () => clearTimeout(timer);
        }
    }, [cartItems, getTotalPrice, totalPrice]);

    return (
        <header className='header'>
            <Link to="/">
                <div className='headerLeft'>
                    <img width={40} height={40} src="/img/logo.svg" alt="logo" />
                    <div className='headerInfo'>
                        <h3>React Sneakers</h3>
                        <p>Магазин лучших кроссовок</p>
                    </div>
                </div>
            </Link>

            <ul className='headerRight'>
                <li onClick={props.onClickCart}>
                    <img width={18} height={18} src="/img/cart.svg" alt="cart" />
                    <span className={priceChanged ? 'price-changed' : ''}>
                        {totalPrice} руб.
                    </span>
                </li>
                <li>
                    <Link to="/favourites">
                        <img width={18} height={18} src="/img/Vector.png" alt="favorite" />
                    </Link>
                </li>
                <li>
                    <Link to="/orders">
                        <img width={18} height={18} src="/img/order.svg" alt="orders" />
                    </Link>
                </li>
                <li>
                    <Link to="/login">
                        <img width={18} height={18} src='/img/user.svg' alt="user" />
                    </Link>
                </li>
            </ul>
        </header>
    );
};