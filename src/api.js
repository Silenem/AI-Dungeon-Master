export const CLASSES = {
  tracker: { id:"tracker", title:"Следопыт", icon:"🏹", description:["Мастер троп.","Видит скрытое."], stats:{str:1,dex:3,int:2,wis:2,cha:0,con:1}, gear:["Лук","Кинжал","Плащ"] },
  warrior: { id:"warrior", title:"Воин", icon:"⚔️", description:["Сталь решает.","Стойкость — второе имя."], stats:{str:3,dex:1,int:0,wis:1,cha:1,con:3}, gear:["Меч","Щит","Кольчуга"] },
  mage: { id:"mage", title:"Маг", icon:"✨", description:["Знание — сила.","Тайны в крови."], stats:{str:0,dex:1,int:3,wis:3,cha:1,con:1}, gear:["Посох","Книга","Мантия"] }
};

export const STAT_LABELS = { str:"Сила", dex:"Ловкость", int:"Интеллект", wis:"Мудрость", cha:"Харизма", con:"Телосложение" };

export async function createCharacter(data) { 
  console.log("createCharacter:", data); 
  return { ok:true, ...data }; 
}

export async function startStory(data) {
  console.log("startStory called with:", data);
  return { ok: true, ...data };
}

export async function getCharacters() {
  return [
    {
      id: 1,
      name: "Эрдан",
      classId: "tracker",
      status: "ready",
      backstory: "Охотник из северных лесов, ищущий следы древнего зверя.",
      currentStory: null
    },
    {
      id: 2,
      name: "Гром",
      classId: "warrior",
      status: "in_adventure",
      backstory: "Бывший стражник, изгнанный из города за неподчинение.",
      currentStory: { title: "Тени подземелья", place: "Заброшенные шахты" }
    },
    {
      id: 3,
      name: "Элара",
      classId: "mage",
      status: "ready",
      backstory: "Ученица академии магии, сбежавшая в поисках запретных знаний.",
      currentStory: null
    }
  ];
}

// --- ФУНКЦИИ ДЛЯ СТРАНИЦЫ ИСТОРИЙ ---

export async function getStory() {
  return {
    id: "dark-forest",
    title: "Тени над Дубравой",
    genres: ["Тёмное фэнтези", "Детектив"],
    paragraphs: [
      "Деревня на опушке леса затихла. Каждую ночь кто-то уносит скот, а на утро в грязи остаются следы, которых не должно быть.",
      "Старейшина просит вас разобраться. Но чем глубже вы заходите в лес, тем яснее понимаете: это не просто зверь."
    ],
    features: ["Диалоги с NPC", "Скрытые улики", "Ветвящийся сюжет"]
  };
}

export async function getProgress() {
  // правая колонка с персонажем
  /*
  return {
    character: { id: 1, name: "Эрдан", classId: "tracker" },
    checkpoint: "Вход в тёмный лес",
    summary: "Вы только что ступили под сень древних деревьев.",
    saved: true
  };
  */
  
  
  return null; 
}