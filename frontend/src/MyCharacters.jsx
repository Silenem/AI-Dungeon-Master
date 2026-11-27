
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CLASSES, STAT_LABELS, getCharacters, startStory } from "./api";
import "./MyCharacters.css";

const ICONS = { guard: "🛡", tracker: "🌲", mystic: "✦" };

export default function MyCharacters() {
  const navigate = useNavigate();
  const [characters, setCharacters] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getCharacters()
      .then(setCharacters)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  async function handleStart(id) {
    try {
      await startStory(id);
      navigate("/stories");
    } catch (e) {
      setError(e.message);
    }
  }

  return (
    <div className="mc-page">
      {/* ----- Шапка ----- */}
      <header className="mc-header">
        <div className="mc-header__left">
          <span className="mc-logo">✦ ИИ-Мастер</span>
          <Link to="/stories" className="mc-nav">Истории</Link>
          <Link to="/characters" className="mc-nav mc-nav--active">Персонажи</Link>
        </div>
        <div className="mc-header__right">
          <button className="mc-link-btn">☼ Светлая тема</button>
          <span>👤 Silenem</span>
        </div>
      </header>

      {/* ----- Контент ----- */}
      <main className="mc-frame">
        <div className="mc-top">
          <div>
            <p className="mc-eyebrow">ВАШИ ГЕРОИ</p>
            <h1 className="mc-title">Мои персонажи</h1>
            <p className="mc-subtitle">Выберите, чья история продолжится сегодня.</p>
          </div>
          <Link to="/create_character" className="mc-btn mc-btn--top">+ Создать персонажа</Link>
        </div>

        {loading && <p className="mc-muted">Загрузка…</p>}
        {error && <p className="mc-error">{error}</p>}

        <div className="mc-grid">
          {characters.map((c) => {
            const cls = CLASSES[c.classId];
            const inAdventure = c.status === "in_adventure";
            return (
              <article className="mc-card" key={c.id}>
                <div className="mc-status">
                  {inAdventure ? "◎ В приключении" : "✓ Готов к приключению"}
                </div>

                <div className="mc-hero">
                  <span className="mc-avatar">{ICONS[c.classId]}</span>
                  <div>
                    <div className="mc-name">{c.name}</div>
                    <div className="mc-muted">{cls.title}</div>
                  </div>
                </div>

                <p className="mc-backstory">{c.backstory}</p>

                <div className="mc-stats">
                  {Object.entries(cls.stats).map(([k, v]) => (
                    <div key={k}>
                      <div className="mc-muted mc-small">{STAT_LABELS[k]}</div>
                      <div className="mc-stats__val">+{v}</div>
                    </div>
                  ))}
                </div>

                <div className="mc-story">
                  <div className="mc-eyebrow mc-small">
                    {inAdventure ? "ТЕКУЩАЯ ИСТОРИЯ" : "ВПЕРЕДИ ПЕРВАЯ ИСТОРИЯ"}
                  </div>
                  <div className="mc-story__title">
                    {inAdventure ? c.currentStory.title : "Путь ещё не выбран"}
                  </div>
                  <div className="mc-muted mc-small">
                    {inAdventure ? c.currentStory.place : "Можно начать приключение"}
                  </div>
                </div>

                <div className="mc-actions">
                  <button className="mc-btn mc-btn--ghost" onClick={() => navigate(`/characters/${c.id}`)}>
                    Подробнее
                  </button>
                  {inAdventure ? (


<button className="mc-btn" onClick={() => navigate("/stories")}>Продолжить →</button>
                  ) : (
                    <button className="mc-btn" onClick={() => handleStart(c.id)}>Начать историю →</button>
                  )}
                </div>
              </article>
            );
          })}
        </div>
      </main>
    </div>
  );
}