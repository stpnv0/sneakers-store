import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from '../api/axios';
import "../index.scss";

// Функция для декодирования JWT токена (используется только внутри компонента)
const decodeJWT = (token) => {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    console.error('Error decoding JWT:', e);
    return null;
  }
};

const Login = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: "", 
    password: "",
    app_id: 1,
  });
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState("");

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleClose = () => {
    if (window.confirm("Вы уверены, что хотите закрыть форму?")) {
      navigate("/");
    }
  };
  
  const verifyToken = async (token) => {
    try {
      // Проверяем токен, делая запрос к защищенному эндпоинту
      const testAxios = axios.create({
        baseURL: 'http://localhost',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      });

      // Проверка доступа к избранным товарам
      await testAxios.get('/api/v1/favourites');
      return true;
    } catch (error) {
      return false;
    }
  };
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
  
    const endpoint = isRegister ? "/api/v1/auth/register" : "/api/v1/auth/login";
  
    try {
      // Выполняем запрос на авторизацию
      const response = await axios.post(endpoint, formData);
      const data = response.data;
      
      if (!isRegister) {
        if (!data.token) {
          throw new Error("Токен не получен от сервера");
        }
        
        // Сохраняем токен
        localStorage.setItem("token", data.token);
        
        // Проверяем сохранение
        const savedToken = localStorage.getItem("token");
        if (savedToken !== data.token) {
          throw new Error("Ошибка сохранения токена");
        }
        
        // Проверяем токен
        const isValid = await verifyToken(savedToken);
        if (!isValid) {
          throw new Error("Ошибка проверки токена");
        }
        
        // Успешный вход
        alert("Вход выполнен!");
        window.location.href = '/';
      } else {
        // Успешная регистрация
        alert("Регистрация успешна!");
        setIsRegister(false); // Переключаем на форму входа
      }
    } catch (err) {
      const errorMessage = err.response?.data?.message || err.message || "Ошибка запроса";
      setError(errorMessage);
      localStorage.removeItem("token");
    }
  };

  return (
    <div className="login-modal">
      <div className="login-container">
        <h2>{isRegister ? "Регистрация" : "Вход"}</h2>
        {error && <p className="error">{error}</p>}
        <form onSubmit={handleSubmit}>
          <input
            type="email" 
            name="email" 
            placeholder="Email"
            value={formData.email}
            onChange={handleChange}
            required
          />
          <input
            type="password"
            name="password"
            placeholder="Пароль"
            value={formData.password}
            onChange={handleChange}
            required
          />
          <button type="submit">{isRegister ? "Зарегистрироваться" : "Войти"}</button>
        </form>
        <p onClick={() => setIsRegister(!isRegister)} className="toggle">
          {isRegister ? "Уже есть аккаунт? Войти" : "Нет аккаунта? Зарегистрироваться"}
        </p>
        <button className="close-btn" onClick={handleClose}>Закрыть</button>
      </div>
    </div>
  );
};

export default Login;