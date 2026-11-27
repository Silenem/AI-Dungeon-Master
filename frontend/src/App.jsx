import { useEffect, useState } from 'react';
import './App.css';

const loginPath = '/auth/index.html?mode=login';

export default function App() {
  const [theme, setTheme] = useState('dark');
  const [profile, setProfile] = useState(null);
  const [error, setError] = useState('');
  const [signingOut, setSigningOut] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    async function loadProfile() {
      try {
        const response = await fetch('/api/me', { signal: controller.signal });
        if (response.status === 401) {
          window.location.replace(loginPath);
          return;
        }
        if (!response.ok) throw new Error('Не удалось загрузить профиль.');
        setProfile(await response.json());
      } catch (failure) {
        if (!controller.signal.aborted) setError(failure.message);
      }
    }
    loadProfile();
    return () => controller.abort();
  }, []);

  async function logout() {
    setSigningOut(true);
    setError('');
    try {
      const response = await fetch('/api/auth/session', { method: 'DELETE' });
      if (!response.ok) throw new Error('Не удалось завершить сессию.');
      window.location.replace(loginPath);
    } catch (failure) {
      setError(failure.message);
      setSigningOut(false);
    }
  }

  if (!profile) {
    return (
      <div className="app-container dark">
        <main className="profile-page">
          {error ? <p role="alert">{error}</p> : <p role="status">Загружаем профиль…</p>}
          {error && <a href={loginPath}>Вернуться ко входу</a>}
        </main>
      </div>
    );
  }

  return (
    <div className={`app-container ${theme}`}>
      <div className="background-layer" />
      <nav className="top-nav">
        <div className="nav-links">
          <span className="logo">✦ ИИ-Мастер</span>
          <a href="#stories">Истории</a>
          <a href="#characters">Персонажи</a>
        </div>
        <div className="nav-actions">
          <button className="theme-toggle" onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
            {theme === 'dark' ? '☀ Светлая тема' : '☾ Тёмная тема'}
          </button>
          <span className="user-tag">👤 {profile.nickname}</span>
        </div>
      </nav>

      <main className="profile-page">
        <header className="profile-header">
          <div className="avatar-large" aria-hidden="true">{profile.nickname.slice(0, 1).toUpperCase()}</div>
          <div className="header-info">
            <span className="label">ПРОФИЛЬ ПУТНИКА</span>
            <h1>{profile.nickname}</h1>
          </div>
        </header>

        <div className="profile-content">
          <div className="left-column">
            <section id="stories" className="continue-adventure">
              <h2>Приключения</h2>
              <p className="description">Прохождения появятся после подключения игрового модуля.</p>
            </section>
            <section id="characters" className="characters-section">
              <h3>Мои персонажи</h3>
              <p className="description">Создание персонажей появится после подключения классов.</p>
            </section>
          </div>

          <div className="right-column">
            <section className="personal-data">
              <h3>Личные данные</h3>
              <div className="input-group">
                <label htmlFor="profile-nickname">Имя в профиле</label>
                <input id="profile-nickname" value={profile.nickname} readOnly />
              </div>
              <div className="input-group">
                <label htmlFor="profile-email">Электронная почта</label>
                <input id="profile-email" type="email" value={profile.email} readOnly />
              </div>
            </section>
            <section className="settings">
              <button type="button" className="btn-secondary" onClick={logout} disabled={signingOut}>
                {signingOut ? 'Выходим…' : 'Выйти из аккаунта'}
              </button>
              {error && <p role="alert">{error}</p>}
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}
