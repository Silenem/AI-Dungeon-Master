import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { saveUser } from "./useUser";
import "./auth.css";

export default function Register() {
  const navigate = useNavigate();
  const [light, setLight] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPass, setShowPass] = useState(false);
  const [error, setError] = useState("");

  const onSubmit = (e) => {
    e.preventDefault();
    if (!name.trim()) return setError("Укажите, как к вам обращаться");
    if (!/^\S+@\S+\.\S+$/.test(email)) return setError("Введите корректную почту");
    if (password.length < 8) return setError("Пароль — не менее 8 символов");
    setError("");
    saveUser({ name: name.trim(), email });
    navigate("/profile");
  };

  return (
    <div className={`rg-page ${light ? "rg-page--light" : ""}`}>
      
      {/* ШАПКА КАК НА ДРУГИХ СТРАНИЦАХ */}
      <header className="ss-header">
        <div className="ss-header__left">
          <Link to="/my_characters" style={{ textDecoration: "none" }}>
            <span className="ss-logo">✦ AI Dungeon Master</span>
          </Link>
          <Link to="/stories" className="ss-nav">Истории</Link>
          <Link to="/my_characters" className="ss-nav">Персонажи</Link>
        </div>
        <div className="ss-header__right">
          <button className="ss-link-btn" onClick={() => setLight(!light)}>
            {light ? "☾ Тёмная тема" : "☼ Светлая тема"}
          </button>
          <span className="ss-nav" style={{ cursor: "default", color: "var(--gold)" }}>
            Регистрация
          </span>
        </div>
      </header>

      <main className="rg-frame">
        <section className="rg-hero">
          <p className="rg-eyebrow">AI DUNGEON MASTER</p>
          <h1 className="rg-title">
            У каждой истории есть первый шаг.
          </h1>
          <p className="rg-lead">
            Создайте героя. Войдите в приключение. Решите, чем оно закончится.
          </p>
          <hr className="rg-divider" />
          <p className="rg-quote">
            Мастер задаёт мир. Вы выбираете свой путь.
          </p>
        </section>

        <form className="rg-card" onSubmit={onSubmit} noValidate>
          <p className="rg-eyebrow">ДОБРО ПОЖАЛОВАТЬ, ПУТНИК</p>
          <h2 className="rg-card__title">Создать аккаунт</h2>

          <label className="rg-label" htmlFor="rg-name">Ваше имя</label>
          <input
            id="rg-name" className="rg-input" type="text"
            placeholder="Как к вам обращаться?"
            value={name} onChange={(e) => setName(e.target.value)}
          />

          <label className="rg-label" htmlFor="rg-email">Электронная почта</label>
          <input
            id="rg-email" className="rg-input" type="email"
            placeholder="you@example.com"
            value={email} onChange={(e) => setEmail(e.target.value)}
          />

          <label className="rg-label" htmlFor="rg-pass">Пароль</label>
          <div className="rg-pass">
            <input
              id="rg-pass" className="rg-input" type={showPass ? "text" : "password"}
              placeholder="Не менее 8 символов"
              value={password} onChange={(e) => setPassword(e.target.value)}
            />
            <button
              type="button" className="rg-eye"
              aria-label={showPass ? "Скрыть пароль" : "Показать пароль"}
              onClick={() => setShowPass(!showPass)}
            >
              {showPass ? "🙈" : "👁"}
            </button>
          </div>

          {error && <p className="rg-error">{error}</p>}

          <button type="submit" className="rg-btn">Создать аккаунт →</button>

          <p className="rg-foot">
            Уже есть аккаунт? <Link to="/auth" className="rg-foot__link">Войти</Link>
          </p>
        </form>
      </main>
    </div>
  );
}