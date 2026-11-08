import React, { useState } from 'react';
import './App.css';

const mockData = {
  Name: "Silenem",
  Email: "player@example.com",
  CurrentStory: {
    Title: "Тень старой башни",
    Location: "У запертой башни · Эрдан, следопыт",
    Description: "Медный ключ уже у вас. За дверью кто-то есть — и он услышал ваши шаги.",
    Health: "16 / 20"
  },
  Characters: [
    { Id: 1, Name: "Эрдан", Class: "Следопыт" },
    { Id: 2, Name: "Бран", Class: "Страж" },
    { Id: 3, Name: "Лиора", Class: "Мистик" }
  ]
};
const CharacterCard = ({ character }) => (
  <div className="character-card">
    {/* Заглушка для иконки (треугольник/круг/звезда) */}
    <div className="char-icon-placeholder">
      {character.Id === 1 ? '▲' : character.Id === 2 ? '⬤' : '✦'}
    </div>
    <div className="char-info">
      <h4>{character.Name}</h4>
      <p>{character.Class}</p>
    </div>
  </div>
);
export default function App() {
  const [theme, setTheme] = useState('dark');

  const toggleTheme = (newTheme) => setTheme(newTheme);

  return (
    <div className={`app-container ${theme}`}>
      {/* Слой для затемнения/осветления фона */}
      <div className="background-layer" />

      {/* Навигация */}
      <nav className="top-nav">
        <div className="nav-links">
          <span className="logo">✦ ИИ-Мастер</span>
          <a href="#stories">Истории</a>
          <a href="#characters">Персонажи</a>
        </div>
        <div className="nav-actions">
          <button onClick={() => toggleTheme(theme === 'dark' ? 'light' : 'dark')} className="theme-toggle">
            {theme === 'dark' ? '☀ Светлая тема' : '☾ Тёмная тема'}
          </button>
          <span className="user-tag">👤 {mockData.Name}</span>
        </div>
      </nav>

      {/* Основной контент */}
      <main className="profile-page">
        <header className="profile-header">
          <div className="avatar-large">S</div>
          <div className="header-info">
            <span className="label">ПРОФИЛЬ ПУТНИКА</span>
            <h1>{mockData.Name}</h1>
            <p>Ваши герои и незаконченные истории.</p>
          </div>
        </header>

        <div className="profile-content">
          {/* Левая колонка */}
          <div className="left-column">
            <section className="continue-adventure">
              <div className="section-header">
                <span className="label">ПРОДОЛЖИТЬ ПРИКЛЮЧЕНИЕ</span>
                <span className="saved-status">✓ Сохранено</span>
              </div>
              <h2>{mockData.CurrentStory.Title}</h2>
              <p className="location">📍 {mockData.CurrentStory.Location}</p>
              <p className="description">{mockData.CurrentStory.Description}</p>
              
              <div className="action-row">
                <span className="health">♡ Здоровье {mockData.CurrentStory.Health}</span>
                <button className="btn-primary">Вернуться в историю →</button>
              </div>
            </section>

            <section className="characters-section">
              <div className="section-header">
                <h3>Мои персонажи</h3>
                <a href="#all" className="link-muted">Все персонажи ↗</a>
              </div>
              <div className="characters-grid">
                {mockData.Characters.map(char => (
                  <CharacterCard key={char.Id} character={char} />
                ))}
              </div>
            </section>
          </div>

          {/* Правая колонка */}
          <div className="right-column">
            <section className="personal-data">
              <h3>Личные данные</h3>
              <div className="input-group">
                <label>Имя в профиле</label>
                <input type="text" defaultValue={mockData.Name} />
              </div>
              <div className="input-group">
                <label>Электронная почта</label>
                <input type="email" defaultValue={mockData.Email} />
              </div>
              <button className="btn-secondary">Сохранить изменения</button>
            </section>

            <section className="settings">
              <h3>Оформление</h3>
              <div className="theme-switcher">
                <button className={`theme-btn ${theme === 'light' ? 'active' : ''}`} onClick={() => toggleTheme('light')}>
                  ☀ Светлое
                </button>
                <button className={`theme-btn ${theme === 'dark' ? 'active' : ''}`} onClick={() => toggleTheme('dark')}>
                  ☾ Тёмное
                </button>
              </div>
              <a href="#logout" className="logout-link">↪ Выйти из аккаунта</a>
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}
