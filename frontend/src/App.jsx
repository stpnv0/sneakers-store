import { useState } from 'react'
import { Route, Routes } from "react-router-dom";
import { Header } from './components/Header'
import { Drawer } from './components/Drawer'
import Home from './pages/Home'
import { Favourites } from './pages/Favourites'
import Orders from './pages/Orders'
import Login from './pages/Login'
import { FavoritesProvider } from './context/FavoritesContext'
import { CartProvider } from './context/CartContext'
import { ItemsProvider } from './context/ItemsContext'

export const App = () => {
  const [searchValue, setSearchValue] = useState('');
  const [cartOpened, setCartOpened] = useState(false);

  const onChangeSearchInput = (event) => {
    setSearchValue(event.target.value);
  }

  return (
    <ItemsProvider>
      <FavoritesProvider>
        <CartProvider>
          <div className='wrapper'>
            {cartOpened && (
              <Drawer onClose={() => setCartOpened(false)} />
            )}
            <Header onClickCart={() => setCartOpened(true)} />
            <Routes>
              <Route path="/"
                element={
                  <Home
                    searchValue={searchValue}
                    setSearchValue={setSearchValue}
                    onChangeSearchInput={onChangeSearchInput}
                  />
                }
              />
              <Route path='/favourites'
                element={<Favourites />}
              />
              <Route path='/orders'
                element={<Orders />}
              />
              <Route path="/login"
                element={<Login />}
              />
            </Routes>
          </div>
        </CartProvider>
      </FavoritesProvider>
    </ItemsProvider>
  )
}

