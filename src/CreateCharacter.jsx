import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CLASSES, STAT_LABELS, createCharacter } from "./api";
import "./CreateCharacter.css";

export default function CreateCharacter() {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [classId, setClassId] = useState("tracker");
  const [backstory, setBackstory] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const hero = CLASSES[classId];

  async function handleSubmit(e) {
    e.preventDefault();
    if (!name.trim()) return setError("Введите имя персонажа");
    setError("");
    setLoading(true);
    try {
      await createCharacter({ name: name.trim(), classId, backstory });
      navigate("/characters");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="cc-page">
      {/* ----- Шапка ----- */}
      <header className="cc-header">
        <div className="ss-header__left">
          <span className="ss-logo">✦ ИИ-Мастер</span>
          <Link to="/stories" className="ss-nav">Истории</Link>
          <Link to="/my_characters" className="ss-nav">Персонажи</Link>
        </div>
        <div className="ss-header__right">
          <button className="ss-link-btn">☼ Светлая тема</button>
          <Link to="#" className="ss-nav">👤 Silenem</Link>
        </div>
      </header>

      {/* ----- Контент ----- */}
      <main className="cc-frame">
        <form className="cc-main" onSubmit={handleSubmit}>
          <p className="cc-eyebrow">ВАШ СЛЕДУЮЩИЙ ГЕРОЙ</p>
          <h1 className="cc-title">Создание персонажа</h1>
          <p className="cc-subtitle">Имя, сильные стороны и прошлое. Всё остальное решит приключение.</p>

          <label className="cc-label" htmlFor="name">Имя персонажа</label>
          <input
            id="name"
            className="cc-input"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Эрдан"
          />

          <h2 className="cc-h2">Выберите свой путь</h2>
          <div className="cc-classes">
            {Object.values(CLASSES).map((c) => (
              <button
                type="button"
                key={c.id}
                className={`cc-class ${classId === c.id ? "cc-class--active" : ""}`}
                onClick={() => setClassId(c.id)}
              >
                <span className="cc-class__icon">{c.icon}</span>
                <span className="cc-class__title">{c.title}</span>
                <span className="cc-class__desc">
                  {c.description[0]}
{c.description[1]}
                </span>
                <span className="cc-class__bonus">
                  {Object.entries(c.stats).filter(([, v]) => v === 3).map(([k]) => `${STAT_LABELS[k]} +3`)}
                </span>
              </button>
            ))}
          </div>

          <label className="cc-label" htmlFor="story">
            Предыстория <span className="cc-muted">· необязательно</span>
          </label>
          <textarea
            id="story"
            className="cc-input cc-textarea"
            rows={3}
            value={backstory}
            onChange={(e) => setBackstory(e.target.value)}
            placeholder="Откуда вы пришли? Что ищете? Какую тайну храните?"
          />

          {error && <p className="cc-error">{error}</p>}

          <div className="cc-actions">
            <button className="cc-btn" type="submit" disabled={loading}>
              {loading ? "Создаём…" : "Создать персонажа →"}
            </button>
            <Link to="/my_characters" className="cc-cancel">Отмена</Link>
          </div>
        </form>

        {/* ----- Лист персонажа ----- */}
        <aside className="cc-sheet">
          <p className="cc-eyebrow">ЛИСТ ПЕРСОНАЖА</p>
          <div className="cc-sheet__hero">
            <span className="cc-sheet__avatar">{hero.icon}</span>
            <div>
              <div className="cc-sheet__name">{name || "Без имени"}</div>
              <div className="cc-muted">{hero.title}</div>
            </div>
          </div>

          <div className="cc-health">
            <div className="cc-health__row"><span>♡ Здоровье</span><b>20 / 20</b></div>
            <div className="cc-health__bar" />
          </div>

          <div className="cc-stats">
            {Object.entries(hero.stats).map(([k, v]) => (
              <div key={k}>
                <div className="cc-muted">{STAT_LABELS[k]}</div>
                <div className="cc-stats__val">+{v}</div>
              </div>
            ))}
          </div>

          <h3 className="cc-sheet__h3">Стартовое снаряжение</h3>
          <ul className="cc-gear">
            {hero.gear.map((g) => <li key={g}>{g}</li>)}
          </ul>
          <p className="cc-note">Путь задаёт начальные умения и вещи. Действия выбираете вы.</p>
        </aside>
      </main>
    </div>
  );
}