export async function postAuth(action, data) {
  const response = await fetch(`/api/auth/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    const failure = await response.json().catch(() => null);
    throw new Error(failure?.error ?? 'Не удалось связаться с сервером.');
  }
  return response.status === 202 ? response.json() : null;
}
