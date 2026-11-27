import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { saveUser } from "./useUser";
import "./auth.css";

export default function Register() {
  const navigate = useNavigate();
  const [light, setLight] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPass, setShowPass] = useState(false);
  const [error, setError] = useState("");

  const onSubmit = (e) => {
    e.preventDefault();
    if (!name.trim()) return setError("Укажите, как к вам обращаться");
    if (!/^\S+@\S+\.\S+$/.test(email)) return setError("Введите корректную почту");
    if (password.length < 8) return setError("Пароль — не менее 8 символов");
    setError("");
    saveUser({ name: name.trim(), email });
    navigate("/profile");
  };

  return (
    <div className={`rg-page ${light ? "rg-page--light" : ""}`}>
      
      {/* ШАПКА КАК НА ДРУГИХ СТРАНИЦАХ */}
      <header className="ss-header">
        <div className="ss-header__left">
          <Link to="/my_characters" style={{ textDecoration: "none" }}>
            <span className="ss-logo">✦ AI Dungeon Master</span>
          </Link>
          <Link to="/stories" className="ss-nav">Истории</Link>
          <Link to="/my_characters" className="ss-nav">Персонажи</Link>
        </div>
        <div className="ss-header__right">
          <button className="ss-link-btn" onClick={() => setLight(!light)}>
            {light ? "☾ Тёмная тема" : "☼ Светлая тема"}
          </button>
          <span className="ss-nav" style={{ cursor: "default", color: "var(--gold)" }}>
            Регистрация
          </span>
        </div>
      </header>
