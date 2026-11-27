
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CLASSES, getStory, getProgress } from "./api";
import "./StorySelect.css";

const FEATURE_ICONS = ["💬", "🔍", "🪧"];

export default function StorySelect() {
  const navigate = useNavigate();
  const [story, setStory] = useState(null);
  const [progress, setProgress] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([getStory(), getProgress()])
      .then(([s, p]) => {
        setStory(s);
        setProgress(p);
      })
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div className="ss-page">
      
      <header className="ss-header">
        <div className="ss-header__left">
          <span className="ss-logo">✦ ИИ-Мастер</span>
          <Link to="/stories" className="ss-nav ss-nav--active">Истории</Link>
          <Link to="/my_characters" className="ss-nav">Персонажи</Link>
        </div>
        <div className="ss-header__right">
          <button className="ss-link-btn">☼ Светлая тема</button>
          <Link to="#" className="ss-nav ">👤 Silenem</Link>
        </div>
      </header>

      {/* ----- Контент ----- */}
      <main className="ss-frame">
        <p className="ss-eyebrow">ПРИКЛЮЧЕНИЯ</p>
        <h1 className="ss-title">Выберите свою историю</h1>
        <p className="ss-subtitle">Мир подготовлен. Каким будет его финал — решать вам.</p>

        {error && <p className="ss-error">{error}</p>}
        {!story && !error && <p className="ss-muted">Загрузка…</p>}

        {story && (
          <div className="ss-layout">
            {/* Карточка истории */}
            <section className="ss-card">
              <div className="ss-card__top">
                <div>
                  <p className="ss-eyebrow">ПОДГОТОВЛЕННОЕ ПРИКЛЮЧЕНИЕ</p>
                  <h2 className="ss-card__title">{story.title}</h2>
                  <div className="ss-tags">
                    {story.genres.map((g) => <span key={g}>{g}</span>)}
                  </div>
                </div>
                <span className="ss-badge">🏰</span>
              </div>

              <div className="ss-text">
                {story.paragraphs.map((p, i) => <p key={i}>{p}</p>)}
              </div>

              <div className="ss-features">
                {story.features.map((f, i) => (
                  <span key={f}>{FEATURE_ICONS[i]} {f}</span>
                ))}
              </div>

              <div className="ss-actions">
                <button className="ss-btn" onClick={() => navigate("/my_characters")}>
                  Выбрать персонажа →
                </button>
                <button className="ss-link-btn" onClick={() => navigate(`/stories/${story.id}`)}>
                  О приключении ›
                </button>
              </div>
            </section>

            {/* Прохождение */}
            <aside className="ss-side">
              <p className="ss-eyebrow">ВАШЕ ПРОХОЖДЕНИЕ</p>
              {progress ? (
                <>
                  <h3 className="ss-side__title">История ждёт вас</h3>
                  <div className="ss-hero">
                    <span className="ss-avatar">🌲</span>
                    <div>
                      <div className="ss-hero__name">{progress.character.name}</div>
                      <div className="ss-muted ss-small">{CLASSES[progress.character.classId].title}</div>
                    </div>
                  </div>

                  <div className="ss-progress">
                    <p className="ss-eyebrow ss-upper">{story.title}</p>
                    <div className="ss-progress__place">{progress.checkpoint}</div>
                    <p className="ss-muted">{progress.summary}</p>
                    {progress.saved && <p className="saved">✓ Прохождение сохранено</p>}
                  </div>

                  <button className="ss-btn ss-btn--full" onClick={() => navigate("/play")}>
                    Продолжить →
                  </button>
                  <button
                    className="ss-link-btn ss-center"
                    onClick={() => navigate(`/characters/${progress.character.id}`)}
                  >
                    Посмотреть героя
                  </button>
                </>
              ) : (
                <>
                  <h3 className="ss-side__title">Пока нет прохождений</h3>
                  <p className="ss-muted">Выберите персонажа, чтобы начать историю.</p>
                </>
              )}
            </aside>
          </div>
        )}
      </main>
    </div>
  );
}