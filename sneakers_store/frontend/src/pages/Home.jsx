import { useContext } from 'react';
import { Card } from '../components/Card/Card';
import { ItemsContext } from '../context/ItemsContext';

const Home = ({ searchValue, onChangeSearchInput }) => {
  const { items, loading, error } = useContext(ItemsContext);

  const renderItems = () => {
    const filteredItems = items.filter(item => 
      item.title.toLowerCase().includes(searchValue.toLowerCase())
    );

    return (
      loading ? (
        <div>Загрузка...</div>
      ) : error ? (
        <div>Ошибка: {error}</div>
      ) : (
        filteredItems.map(item => (
          <Card
            key={item.id}
            id={item.id}
            title={item.title}
            price={item.price}
            imgUrl={item.imageUrl}
            description={item.description}
          />
        ))
      )
    );
  };

  return (
    <div className='content'>
      <div className='contentHeader'>
        <h1>{searchValue ? `Поиск по запросу: "${searchValue}"` : 'Все кроссовки'}</h1>
        <div className='search-block'>
          <img src='/img/search.svg' alt='Search'/>
          <input onChange={onChangeSearchInput} value={searchValue} placeholder='Поиск...'/>
        </div>
      </div>
      <div className='sneakers'>
        {renderItems()}
      </div>
    </div>
  );
};

export default Home;