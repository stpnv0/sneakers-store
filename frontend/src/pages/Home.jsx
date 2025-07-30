import { useContext } from 'react';
import { Card } from '../components/Card/Card';
import { ItemsContext } from '../context/ItemsContext';

const S3_BASE_URL = 'http://localhost:9000/sneakers';

const Home = ({ searchValue, onChangeSearchInput }) => {
  const { items, loading, error } = useContext(ItemsContext);

  const renderItems = () => {
    const filteredItems = items.filter(item => 
      item.title.toLowerCase().includes(searchValue.toLowerCase())
    );

    if (loading) {
      return <div>Загрузка...</div>;
    }

    if (error) {
      return <div>Ошибка: {error}</div>;
    }

    return (
      filteredItems.map(item => {
        const fullImageUrl = item.image_key 
          ? `${S3_BASE_URL}/${item.image_key}` 
          : '/img/placeholder.svg'; 

        return (
          <Card
            key={item.id}
            id={item.id}
            title={item.title}
            price={item.price}
            imgUrl={fullImageUrl}
            description={item.description} 
          />
        );
      })
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