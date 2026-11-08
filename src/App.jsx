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
