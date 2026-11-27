import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CLASSES, getProgress } from "./api";
import { useUser, saveUser, clearUser } from "./useUser";
import "./Profile.css";

const CLASS_ICONS = { tracker: "🌲", warrior: "🛡", mage: "✨" };

const DEMO_CHARACTERS = [
  { id: 1, name: "Эрдан", classId: "tracker", title: "Следопыт" },
  { id: 2, name: "Гром", classId: "warrior", title: "Воин" },
  { id: 3, name: "Элара", classId: "mage", title: "Маг" },
];

export default function Profile() {
  const navigate = useNavigate();
  const user = useUser();
  
  const [progress, setProgress] = useState(null);
  const [error, setError] = useState("");
  const [light, setLight] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [saved, setSaved] = useState(false);

  
  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email || "");
    } else {
      navigate("/auth", { replace: true });
    }
  }, [user, navigate]);

  useEffect(() => {
    getProgress()
      .then(setProgress)
      .catch((e) => setError(e.message));
  }, []);

 
  if (!user) return null;

  const onSave = () => {
  
    saveUser({ name: name.trim() || "Путник", email });
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const onLogout = () => {
    clearUser();
    navigate("/auth", { replace: true });
  };

  const classTitle = (c) =>
    c.title || (CLASSES && CLASSES[c.classId] && CLASSES[c.classId].title) || "";

  return (
    <div className={`pf-page ${light ? "pf-page--light" : ""}`}>
      <header className="pf-header">
        <div className="pf-header__left">
          <span className="pf-logo">✦ AI Dungeon Master</span>
          <Link to="/stories" className="pf-nav">Истории</Link>
          <Link to="/my_characters" className="pf-nav">Персонажи</Link>
        </div>
        <div className="pf-header__right">
          <button className="pf-link-btn" onClick={() => setLight(!light)}>
            {light ? "☾ Тёмная тема" : "☼ Светлая тема"}
          </button>
  
          <Link to="/profile" className="pf-user">👤 {user.name}</Link>
        </div>
      </header>

      <main className="pf-frame">
        <div className="pf-top">
    
          <div className="pf-avatar">{user.name.trim().charAt(0).toUpperCase() || "?"}</div>
          <div>
            <p className="pf-eyebrow">ПРОФИЛЬ ПУТНИКА</p>
            <h1 className="pf-title">{user.name}</h1>
            <p className="pf-subtitle">Ваши герои и незаконченные истории.</p>
          </div>
        </div>

        <div className="pf-layout">
          <div className="pf-main">
            <section className="pf-card">
              <div className="pf-card__head">
                <p className="pf-eyebrow">ПРОДОЛЖИТЬ ПРИКЛЮЧЕНИЕ</p>
                {progress && progress.saved && <span className="pf-saved">✓ Сохранено</span>}
              </div>

              {error && <p className="pf-error">{error}</p>}
              {!progress && !error && <p className="pf-muted">Загрузка…</p>}

              {progress && (
                <>
                  <h2 className="pf-card__title">
                    {progress.storyTitle || "Тень старой башни"}
                  </h2>
                  <p className="pf-meta">
                    <span>⌖ {progress.checkpoint}</span>
                    <span className="pf-dot">·</span>
                    <span>
                      {progress.character.name}
                      {CLASSES && CLASSES[progress.character.classId]
                        ? `, ${CLASSES[progress.character.classId].title.toLowerCase()}`
                        : ""}
                    </span>
                  </p>
                  <p className="pf-summary">{progress.summary}</p>

                  <div className="pf-card__foot">
                    <span className="pf-health">
                      ♡ Здоровье {progress.hp ?? 16} / {progress.maxHp ?? 20}
                    </span>
                    <button


className="pf-btn" onClick={() => navigate("/stories")}>
                      Вернуться в историю →
                    </button>
                  </div>
                </>
              )}
            </section>

            <div className="pf-chars-head">
              <h2 className="pf-h2">Мои персонажи</h2>
              <Link to="/my_characters" className="pf-link">Все персонажи ↗</Link>
            </div>
            <div className="pf-chars">
              {DEMO_CHARACTERS.map((c) => (
                <button key={c.id} className="pf-char" onClick={() => navigate("/my_characters")}>
                  <span className="pf-char__icon">{CLASS_ICONS[c.classId] || "✦"}</span>
                  <span className="pf-char__info">
                    <span className="pf-char__name">{c.name}</span>
                    <span className="pf-char__class">{classTitle(c)}</span>
                  </span>
                  <span className="pf-char__arrow">›</span>
                </button>
              ))}
            </div>
          </div>

          <aside className="pf-side">
            <h2 className="pf-h2">Личные данные</h2>

            <label className="pf-label" htmlFor="pf-name">Имя в профиле</label>
            <input
              id="pf-name" 
              className="pf-input" 
              value={name}
              onChange={(e) => setName(e.target.value)}
            />

            <label className="pf-label" htmlFor="pf-email">Электронная почта</label>
            <input
              id="pf-email" 
              className="pf-input" 
              type="email" 
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />

            <button className="pf-btn pf-btn--ghost" onClick={onSave}>
              {saved ? "✓ Сохранено" : "Сохранить изменения"}
            </button>

            <hr className="pf-hr" />

            <h3 className="pf-side__sub">Оформление</h3>
            <div className="pf-theme">
              <button
                className={`pf-theme__btn ${!light ? "" : "pf-theme__btn--on"}`}
                onClick={() => setLight(true)}
              >
                ☼ Светлое
              </button>
              <button
                className={`pf-theme__btn ${light ? "" : "pf-theme__btn--on"}`}
                onClick={() => setLight(false)}
              >
                ☾ Тёмное
              </button>
            </div>

            <button className="pf-logout" onClick={onLogout}>⇤ Выйти из аккаунта</button>
          </aside>
        </div>
      </main>
    </div>
  );
}