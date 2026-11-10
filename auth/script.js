const DungeonMasterAPI = {
  BASE_URL: 'https://jsonplaceholder.typicode.com',

  async authenticateUser(payload, isLoginMode) {
    const endpoint = isLoginMode ? '/posts' : '/users';
    const response = await fetch(`${this.BASE_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json; charset=UTF-8',
      },
      body: JSON.stringify({
        name: payload.username || 'Путник',
        email: payload.email,
        action: isLoginMode ? 'login' : 'register',
        timestamp: new Date().toISOString(),
      }),
    });

    if (!response.ok) {
      throw new Error(`Ошибка сервера: ${response.status}`);
    }

    const data = await response.json();

    const session = {
      userId: data.id || Math.floor(Math.random() * 1000),
      name: payload.username || payload.email.split('@')[0],
      email: payload.email,
      token: `dm_jwt_${Math.random().toString(36).substring(2)}`,
    };
    localStorage.setItem('dm_session', JSON.stringify(session));

    return session;
  },

  async fetchStartingQuest(heroName) {
    const response = await fetch(`${this.BASE_URL}/posts/1`);
    if (!response.ok) {
      throw new Error('Не удалось связаться с сервером');
    }

    return 'Вы успешно зарегистрировались!';
  },
};


document.addEventListener('DOMContentLoaded', () => {
  const authForm = document.getElementById('authForm');
  const nameGroup = document.getElementById('nameGroup');
  const usernameInput = document.getElementById('username');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');
  const togglePasswordBtn = document.getElementById('togglePassword');

  const cardEyebrow = document.getElementById('cardEyebrow');
  const cardTitle = document.getElementById('cardTitle');
  const submitBtn = document.getElementById('submitBtn');
  const submitBtnText = document.getElementById('submitBtnText');
  const switchPrompt = document.getElementById('switchPrompt');
  const switchModeBtn = document.getElementById('switchModeBtn');
  const headerAuthBtn = document.getElementById('headerAuthBtn');
  const apiStatus = document.getElementById('apiStatus');

  const themeToggle = document.getElementById('themeToggle');
  const themeText = document.getElementById('themeText');
  const themeIcon = document.getElementById('themeIcon');

  let isLoginMode = false;

  // Переключение темы (Тёмная / Светлая)
  themeToggle.addEventListener('click', () => {
    const isLight = document.body.classList.toggle('light-theme');
    themeText.textContent = isLight ? 'Тёмная тема' : 'Светлая тема';
    themeIcon.textContent = isLight ? '🌙' : '☀️';
  });

  // Показать / скрыть пароль
  togglePasswordBtn.addEventListener('click', () => {
    const isPassword = passwordInput.type === 'password';
    passwordInput.type = isPassword ? 'text' : 'password';
  });

  // Переключение режима «Создать аккаунт» / «Войти»
  function toggleAuthMode(forceLogin) {
    isLoginMode = typeof forceLogin === 'boolean' ? forceLogin : !isLoginMode;
    apiStatus.hidden = true;

    if (isLoginMode) {
      nameGroup.style.display = 'none';
      usernameInput.removeAttribute('required');
      cardEyebrow.textContent = 'С ВОЗВРАЩЕНИЕМ, ГЕРОЙ';
      cardTitle.textContent = 'Войти в мир';
      submitBtnText.textContent = 'Войти';
      switchPrompt.textContent = 'Ещё нет аккаунта?';
      switchModeBtn.textContent = 'Создать';
    } else {
      nameGroup.style.display = 'flex';
      usernameInput.setAttribute('required', 'true');
      cardEyebrow.textContent = 'ДОБРО ПОЖАЛОВАТЬ, ПУТНИК';
      cardTitle.textContent = 'Создать аккаунт';
      submitBtnText.textContent = 'Создать аккаунт';
      switchPrompt.textContent = 'Уже есть аккаунт?';
      switchModeBtn.textContent = 'Войти';
    }
  }

  switchModeBtn.addEventListener('click', () => toggleAuthMode());
  headerAuthBtn.addEventListener('click', () => toggleAuthMode(true));

  // Отправка формы через API
  authForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    apiStatus.hidden = true;
    apiStatus.classList.remove('error');

    const username = usernameInput.value.trim();
    const email = emailInput.value.trim();
    const password = passwordInput.value;

    // Простая валидация
    if (!isLoginMode && username.length < 2) {
      showStatus('Пожалуйста, укажите имя вашего героя (минимум 2 символа).', true);
      return;
    }
    if (!email.includes('@')) {
      showStatus('Введите корректный адрес электронной почты.', true);
      return;
    }
    if (password.length < 8) {
      showStatus('Пароль должен содержать не менее 8 символов.', true);
      return;
    }

    submitBtn.disabled = true;
    submitBtnText.textContent = 'Связь с Мастером...';

    try {
      const session = await DungeonMasterAPI.authenticateUser(
        { username, email, password },
        isLoginMode
      );

      const prologue = await DungeonMasterAPI.fetchStartingQuest(session.name);

      showStatus(`✓ Аккаунт готов (ID: #${session.userId}). ${prologue}`, false);
      authForm.reset();
    } catch (error) {
      showStatus(`Ошибка соединения с API: ${error.message}`, true);
    } finally {
      submitBtn.disabled = false;
      submitBtnText.textContent = isLoginMode ? 'Войти' : 'Создать аккаунт';
    }
  });

  function showStatus(message, isError) {
    apiStatus.textContent = message;
    apiStatus.classList.toggle('error', isError);
    apiStatus.hidden = false;
  }
});