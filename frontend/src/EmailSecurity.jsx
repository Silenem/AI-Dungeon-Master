import { useState } from 'react';
import { postAuth } from './auth-api.js';

export default function EmailSecurity({ enabled, onChange }) {
  const [editing, setEditing] = useState(false);
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [challenge, setChallenge] = useState(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  function reset() {
    setEditing(false);
    setPassword('');
    setCode('');
    setChallenge(null);
    setError('');
  }

  async function submit(event) {
    event.preventDefault();
    setError('');
    setBusy(true);
    try {
      if (challenge) {
        await postAuth('2fa/verify', { challenge_id: challenge.challenge_id, code: code.trim() });
      } else {
        const next = await postAuth('2fa', { enabled: !enabled, password });
        if (next) {
          setPassword('');
          setChallenge(next);
          return;
        }
      }
      onChange(!enabled);
      reset();
    } catch (failure) {
      setError(failure.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="settings">
      <h3>Двухфакторный вход</h3>
      <p className="description">{enabled ? 'Код из письма требуется при входе.' : 'Двухфакторный вход выключен.'}</p>
      {editing ? (
        <form onSubmit={submit}>
          {challenge ? (
            <div className="input-group">
              <label htmlFor="two-factor-code">Код из письма</label>
              <input id="two-factor-code" inputMode="numeric" pattern="[0-9]{6}" maxLength={6}
                autoComplete="one-time-code" value={code} onChange={event => setCode(event.target.value)} required />
            </div>
          ) : (
            <div className="input-group">
              <label htmlFor="two-factor-password">Текущий пароль</label>
              <input id="two-factor-password" type="password" autoComplete="current-password"
                value={password} onChange={event => setPassword(event.target.value)} required />
            </div>
          )}
          <button type="submit" className="btn-secondary" disabled={busy}>
            {challenge ? 'Подтвердить код' : 'Отправить код на почту'}
          </button>
          <button type="button" className="btn-secondary" onClick={reset} disabled={busy}>Отмена</button>
        </form>
      ) : (
        <button type="button" className="btn-secondary" onClick={() => setEditing(true)}>
          {enabled ? 'Выключить' : 'Включить'}
        </button>
      )}
      {error && <p role="alert">{error}</p>}
    </section>
  );
}
