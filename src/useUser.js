import { useEffect, useState } from "react";

const KEY = "ai_master_user";

export function getStoredUser() {
  try {
    const raw = localStorage.getItem(KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function saveUser(user) {
  localStorage.setItem(KEY, JSON.stringify(user));
  window.dispatchEvent(new Event("user-changed"));
}

export function clearUser() {
  localStorage.removeItem(KEY);
  window.dispatchEvent(new Event("user-changed"));
}

export function useUser() {
  const [user, setUser] = useState(getStoredUser());

  useEffect(() => {
    const onChange = () => setUser(getStoredUser());
    window.addEventListener("user-changed", onChange);
    window.addEventListener("storage", onChange);
    return () => {
      window.removeEventListener("user-changed", onChange);
      window.removeEventListener("storage", onChange);
    };
  }, []);

  return user;
}
