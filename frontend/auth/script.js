import { postAuth } from '../src/auth-api.js';

document.addEventListener('DOMContentLoaded', () => {
  const authForm = document.getElementById('authForm');
  const nameGroup = document.getElementById('nameGroup');
  const emailGroup = document.getElementById('emailGroup');
  const passwordGroup = document.getElementById('passwordGroup');
  const codeGroup = document.getElementById('codeGroup');
  const usernameInput = document.getElementById('username');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');
  const codeInput = document.getElementById('emailCode');
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

  let isLoginMode = new URLSearchParams(window.location.search).get('mode') === 'login';
  let challenge = null;

  themeToggle.addEventListener('click', () => {
    const isLight = document.body.classList.toggle('light-theme');
    themeText.textContent = isLight ? 'Тёмная тема' : 'Светлая тема';
    themeIcon.textContent = isLight ? '🌙' : '☀️';
  });

  togglePasswordBtn.addEventListener('click', () => {
    passwordInput.type = passwordInput.type === 'password' ? 'text' : 'password';
  });

  function renderForm() {
    const awaitingCode = Boolean(challenge);
    nameGroup.hidden = isLoginMode || awaitingCode;
    emailGroup.hidden = awaitingCode;
    passwordGroup.hidden = awaitingCode;
    codeGroup.hidden = !awaitingCode;
    passwordInput.autocomplete = isLoginMode ? 'current-password' : 'new-password';
    cardEyebrow.textContent = isLoginMode ? 'С ВОЗВРАЩЕНИЕМ, ГЕРОЙ' : 'ДОБРО ПОЖАЛОВАТЬ, ПУТНИК';
    cardTitle.textContent = awaitingCode ? 'Подтвердите почту' : isLoginMode ? 'Войти в мир' : 'Создать аккаунт';
    submitBtnText.textContent = awaitingCode ? 'Подтвердить код' : isLoginMode ? 'Войти' : 'Создать аккаунт';
    switchPrompt.hidden = awaitingCode;
    switchPrompt.textContent = isLoginMode ? 'Ещё нет аккаунта?' : 'Уже есть аккаунт?';
    switchModeBtn.textContent = awaitingCode ? 'Начать заново' : isLoginMode ? 'Создать' : 'Войти';
  }

  function resetForm(loginMode) {
    isLoginMode = loginMode;
    challenge = null;
    passwordInput.value = '';
    codeInput.value = '';
    apiStatus.hidden = true;
    renderForm();
  }

  switchModeBtn.addEventListener('click', () => {
    resetForm(challenge ? isLoginMode : !isLoginMode);
    emailInput.focus();
  });
  headerAuthBtn.addEventListener('click', () => resetForm(true));
  renderForm();

  authForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    if (submitBtn.disabled) return;
    apiStatus.hidden = true;
    const action = isLoginMode ? 'login' : 'register';
    const username = usernameInput.value.trim();
    const email = emailInput.value.trim();
    const password = passwordInput.value;
    const code = codeInput.value.trim();

    if (challenge && !/^\d{6}$/.test(code)) {
      showStatus('Введите шестизначный код из письма.', true);
      return;
    }
    if (!challenge && !validateCredentials(username, email, password)) return;

    submitBtn.disabled = true;
    switchModeBtn.disabled = true;
    headerAuthBtn.disabled = true;
    submitBtnText.textContent = 'Подождите…';
    try {
      if (challenge) {
        await postAuth(`${action}/verify`, { challenge_id: challenge.challenge_id, code });
        if (!isLoginMode) {
          resetForm(true);
          passwordInput.focus();
          showStatus('Почта подтверждена. Войдите с вашим паролем.', false);
          return;
        }
      } else {
        challenge = await postAuth(action, { email, password, ...(isLoginMode ? {} : { nickname: username }) });
        if (challenge) {
          passwordInput.value = '';
          renderForm();
          codeInput.focus();
          showStatus(`Код отправлен на ${email}. Он действует 10 минут.`, false);
          return;
        }
      }
      window.location.assign('/');
    } catch (error) {
      showStatus(error.message, true);
    } finally {
      submitBtn.disabled = false;
      switchModeBtn.disabled = false;
      headerAuthBtn.disabled = false;
      renderForm();
    }
  });

  function validateCredentials(username, email, password) {
    if (!isLoginMode && username.length < 2) {
      showStatus('Укажите имя длиной от 2 символов.', true);
      return false;
    }
    if (!email || !emailInput.validity.valid) {
      showStatus('Введите корректный адрес электронной почты.', true);
      return false;
    }
    if ([...password].length < 8) {
      showStatus('Пароль должен содержать не менее 8 символов.', true);
      return false;
    }
    return true;
  }

  function showStatus(message, isError) {
    apiStatus.textContent = message;
    apiStatus.classList.toggle('error', isError);
    apiStatus.hidden = false;
  }
});
