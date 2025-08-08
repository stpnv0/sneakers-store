import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "../api/axios";
import styles from "./Login.module.scss";

const Login = () => {
  const navigate = useNavigate();

  const [formData, setFormData] = useState({
    email: "",
    password: "",
    app_id: 1,
  });
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const toggleShowPassword = () => setShowPassword((s) => !s);

  const handleChange = (e) =>
    setFormData({ ...formData, [e.target.name]: e.target.value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");

    const endpoint = isRegister ? "/api/v1/auth/register" : "/api/v1/auth/login";
    const payload = isRegister
      ? { email: formData.email, password: formData.password }
      : formData;

    try {
      const { data } = await axios.post(endpoint, payload);

      if (!isRegister) {
        const { token } = data;
        if (!token) throw new Error("Токен не был получен от сервера.");
        localStorage.setItem("token", token);
        window.location.href = "/";
      } else {
        alert("Регистрация успешна! Теперь вы можете войти.");
        setIsRegister(false);
        setFormData((prev) => ({ ...prev, password: "" }));
      }
    } catch (err) {
      const msg = err.response?.data?.error || err.message || "Произошла ошибка";
      setError(msg);
      localStorage.removeItem("token");
    }
  };

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h2 className={styles.title}>
          {isRegister ? "Регистрация" : "Вход"}
        </h2>

        {error && <div className={styles.error}>{error}</div>}

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              name="email"
              placeholder="you@example.com"
              value={formData.email}
              onChange={handleChange}
              required
              autoFocus
              autoComplete="email"
            />
          </div>

          <div className={styles.field}>
            <label htmlFor="password">Пароль</label>
            <div className={styles.passwordInput}>
              <input
                id="password"
                type={showPassword ? "text" : "password"}
                name="password"
                placeholder="••••••••"
                value={formData.password}
                onChange={handleChange}
                required
                autoComplete={isRegister ? "new-password" : "current-password"}
              />
              <button
                type="button"
                onClick={toggleShowPassword}
                className={styles.eyeBtn}
                aria-label={showPassword ? "Скрыть пароль" : "Показать пароль"}
                aria-pressed={showPassword}
              >
                {showPassword ? (
                  // eye-off
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M2.1 3.51 3.51 2.1 21.9 20.49 20.49 21.9l-2.3-2.3A11.6 11.6 0 0 1 12 20c-5.5 0-9.5-3.7-11-8 1-2.8 2.9-5.1 5.3-6.6L2.1 3.51zM7.2 8.61A7.9 7.9 0 0 0 3 12c1.2 3 4.6 6 9 6 1.7 0 3.2-.4 4.6-1.2l-2.2-2.2a4 4 0 0 1-5.2-5.2L7.2 8.6zm5.5-.8a4 4 0 0 1 4.4 4.4l-1.7-1.7a2 2 0 0 0-1-1l-1.7-1.7zM12 6c5.5 0 9.5 3.7 11 8-.7 2-2 3.8-3.6 5.2l-1.4-1.4C19.4 16.4 21 14.4 21 12c-1.2-3-4.6-6-9-6-1.1 0-2.2.2-3.2.5L7.6 5.3A11.6 11.6 0 0 1 12 6z"/>
                  </svg>
                ) : (
                  // eye
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 4.5c5.5 0 9.5 3.7 11 8-1.5 4.3-5.5 8-11 8s-9.5-3.7-11-8c1.5-4.3 5.5-8 11-8zm0 2c-4.4 0-7.8 3-9 6 1.2 3 4.6 6 9 6s7.8-3 9-6c-1.2-3-4.6-6-9-6zm0 2.5a3.5 3.5 0 1 1 0 7 3.5 3.5 0 0 1 0-7z"/>
                  </svg>
                )}
              </button>
            </div>
          </div>

          <button type="submit" className={styles.submit}>
            {isRegister ? "Зарегистрироваться" : "Войти"}
          </button>

          <div className={styles.helperRow}>
            <button
              type="button"
              className={styles.linkBtn}
              onClick={() => setIsRegister(!isRegister)}
            >
              {isRegister ? "Уже есть аккаунт? Войти" : "Нет аккаунта? Зарегистрироваться"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default Login;