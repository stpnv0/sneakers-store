import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from '../api/axios';
import "../index.scss";

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
    navigate("/");
  };
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
  
    const endpoint = isRegister ? "/api/v1/auth/register" : "/api/v1/auth/login";
    
    const payload = isRegister 
      ? { email: formData.email, password: formData.password } 
      : formData;

    try {
      const response = await axios.post(endpoint, payload);
      
      if (!isRegister) {
        const { token } = response.data;
        if (!token) {
          throw new Error("Токен не был получен от сервера.");
        }
        
        localStorage.setItem("token", token);
        alert("Вход выполнен успешно!");
        window.location.href = '/'; 
        // ------------------------------------
      } else {
        alert("Регистрация успешна! Теперь вы можете войти.");
        setIsRegister(false); 
        setFormData({ ...formData, password: "" }); 
      }
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.message || "Произошла ошибка";
      console.error("Login/Register error:", err.response || err);
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