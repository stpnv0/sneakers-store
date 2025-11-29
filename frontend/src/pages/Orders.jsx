import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from '../api/axios';
import styles from './Orders.module.scss';

const S3_BASE_URL = 'http://localhost:9000/sneakers';

const Orders = () => {
    const [orders, setOrders] = useState([]);
    const [products, setProducts] = useState({});
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [selectedOrder, setSelectedOrder] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        const fetchOrders = async () => {
            try {
                const token = localStorage.getItem('token');
                if (!token) {
                    navigate('/login');
                    return;
                }

                // Fetch orders
                const ordersResponse = await axios.get('/api/v1/orders');
                const ordersData = ordersResponse.data || [];
                setOrders(ordersData);

                // Collect all unique sneaker IDs
                const sneakerIds = new Set();
                ordersData.forEach(order => {
                    order.items?.forEach(item => {
                        sneakerIds.add(item.sneaker_id);
                    });
                });

                // Fetch product details
                if (sneakerIds.size > 0) {
                    const idsParam = Array.from(sneakerIds).join(',');
                    const productsResponse = await axios.get(`/api/v1/products/batch?ids=${idsParam}`);
                    const productsMap = {};
                    productsResponse.data.forEach(product => {
                        productsMap[product.id] = product;
                    });
                    setProducts(productsMap);
                }
            } catch (err) {
                console.error('Error fetching orders:', err);
                setError(err.response?.data?.error || 'Failed to load orders');
                if (err.response?.status === 401) {
                    navigate('/login');
                }
            } finally {
                setLoading(false);
            }
        };

        fetchOrders();
    }, [navigate]);

    const getStatusBadge = (status) => {
        const statusMap = {
            PENDING_PAYMENT: { text: 'Ожидает оплаты', className: styles.statusPending },
            PAID: { text: 'Оплачен', className: styles.statusPaid },
            PAYMENT_FAILED: { text: 'Ошибка оплаты', className: styles.statusCancelled },
            CANCELLED: { text: 'Отменен', className: styles.statusCancelled }
        };
        const statusInfo = statusMap[status] || { text: status, className: styles.statusDefault };
        return <span className={`${styles.statusBadge} ${statusInfo.className}`}>{statusInfo.text}</span>;
    };

    const formatDate = (timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    const handleOrderClick = (order) => {
        if (order.status === 'PENDING_PAYMENT' && order.payment_url) {
            window.open(order.payment_url, '_blank');
        }
    };

    if (loading) {
        return <div className={styles.loading}>Загрузка заказов...</div>;
    }

    if (error) {
        return <div className={styles.error}>{error}</div>;
    }

    return (
        <div className={styles.ordersContainer}>
            <h1>Мои заказы</h1>
            {orders.length === 0 ? (
                <div className={styles.emptyState}>
                    <img src="/img/empty-box.svg" alt="No orders" />
                    <p>У вас пока нет заказов</p>
                </div>
            ) : (
                <div className={styles.ordersList}>
                    {orders.map((order) => (
                        <div
                            key={order.id}
                            className={`${styles.orderCard} ${order.status === 'PENDING_PAYMENT' ? styles.clickable : ''}`}
                        >
                            <div className={styles.orderHeader}>
                                <div className={styles.orderInfo}>
                                    <h3>Заказ #{order.id}</h3>
                                    <p className={styles.orderDate}>{formatDate(order.created_at)}</p>
                                </div>
                                {getStatusBadge(order.status)}
                            </div>
                            <div className={styles.orderItems}>
                                {order.items?.map((item, index) => {
                                    const product = products[item.sneaker_id];
                                    const imageUrl = product?.image_key
                                        ? `${S3_BASE_URL}/${product.image_key}`
                                        : '/img/placeholder.svg';

                                    return (
                                        <div key={index} className={styles.orderItem}>
                                            <img
                                                src={imageUrl}
                                                alt={product?.title || 'Product'}
                                                className={styles.itemImage}
                                            />
                                            <div className={styles.itemInfo}>
                                                <p className={styles.itemTitle}>
                                                    {product?.title || `Товар #${item.sneaker_id}`}
                                                </p>
                                                <p className={styles.itemQuantity}>Количество: {item.quantity}</p>
                                            </div>
                                            <p className={styles.itemPrice}>{item.price_at_purchase} ₽</p>
                                        </div>
                                    );
                                })}
                            </div>
                            <div className={styles.orderTotal}>
                                <strong>Итого: </strong>
                                <span>{order.total_amount} ₽</span>
                            </div>
                            {order.status === 'PENDING_PAYMENT' && order.payment_url && (
                                <div
                                    className={styles.paymentHint}
                                    onClick={() => handleOrderClick(order)}
                                >
                                    <span>Оплатить</span>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default Orders;
