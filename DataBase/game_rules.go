package domain

// --- ТИПЫ ДОСПЕХОВ И РЕЖИМЫ 0 HP ---

type ArmorType string

const (
	ArmorNone   ArmorType = "none"
	ArmorLight  ArmorType = "light"
	ArmorMedium ArmorType = "medium"
	ArmorHeavy  ArmorType = "heavy"
)

type ZeroHPMode string

const (
	ZeroHPUnconscious ZeroHPMode = "unconscious" // Потеря сознания (нарративное поражение)
	ZeroHPDeathSaves  ZeroHPMode = "death_saves" // Спасброски от смерти
	ZeroHPDead        ZeroHPMode = "dead"        // Мгновенная смерть
)

// --- ПРАВИЛА КЛАССОВ И ПОДКЛАССОВ (Дерево классов v0) ---

type ClassRule struct {
	ID            string
	HitDie        int // d10, d8, d6
	BaseAC        int // Стартовый КД с базовым снаряжением
	PrimaryStat   string
	IsSpellcaster bool
}

type SubclassRule struct {
	ID            string
	ParentClassID string
	AllowedArmor  []ArmorType
	CanUseShield  bool
}

var ClassRules = map[string]ClassRule{
	ClassFighter: {
		ID:            ClassFighter,
		HitDie:        10, // 10 + мод. Телосложения (+1) = 11 хитов на 1 ур.
		BaseAC:        14,
		PrimaryStat:   "strength",
		IsSpellcaster: false,
	},
	ClassRanger: {
		ID:            ClassRanger,
		HitDie:        8, // 8 + мод. Телосложения (+1) = 9 хитов на 1 ур.
		BaseAC:        13,
		PrimaryStat:   "dexterity",
		IsSpellcaster: false,
	},
	ClassMystic: {
		ID:            ClassMystic,
		HitDie:        6, // 6 + мод. Телосложения (+1) = 7 хитов на 1 ур.
		BaseAC:        10,
		PrimaryStat:   "willpower",
		IsSpellcaster: true,
	},
}

var SubclassRules = map[string]SubclassRule{
	SubclassGuardian: {
		ID:            SubclassGuardian,
		ParentClassID: ClassFighter,
		AllowedArmor:  []ArmorType{ArmorNone, ArmorLight, ArmorMedium, ArmorHeavy},
		CanUseShield:  true,
	},
	SubclassSlayer: {
		ID:            SubclassSlayer,
		ParentClassID: ClassFighter,
		AllowedArmor:  []ArmorType{ArmorNone, ArmorLight, ArmorMedium},
		CanUseShield:  false,
	},
	SubclassRogue: {
		ID:            SubclassRogue,
		ParentClassID: ClassRanger,
		AllowedArmor:  []ArmorType{ArmorNone, ArmorLight},
		CanUseShield:  false,
	},
	SubclassHunter: {
		ID:            SubclassHunter,
		ParentClassID: ClassRanger,
		AllowedArmor:  []ArmorType{ArmorNone, ArmorLight, ArmorMedium},
		CanUseShield:  true,
	},
	SubclassWizard: {
		ID:            SubclassWizard,
		ParentClassID: ClassMystic,
		AllowedArmor:  []ArmorType{ArmorNone},
		CanUseShield:  false,
	},
	SubclassSorcerer: {
		ID:            SubclassSorcerer,
		ParentClassID: ClassMystic,
		AllowedArmor:  []ArmorType{ArmorNone},
		CanUseShield:  false,
	},
}

// --- РАСЧЁТ БОЕВЫХ ПАРАМЕТРОВ ПЕРСОНАЖА ---

// MaxHP рассчитывает максимум хитов: 1-й уровень = кость хитов + мод. Телосложения,
// каждый следующий уровень = (кость/2 + 1) + мод. Телосложения (минимум +1 за уровень).
func (c *Character) MaxHP() int {
	rule, ok := ClassRules[c.ClassID]
	if !ok {
		return 1
	}

	conMod := c.Stats.Modifier(c.Stats.Constitution)
	hp := max(1, rule.HitDie+conMod)

	if c.Level > 1 {
		perLevel := max(1, (rule.HitDie/2+1)+conMod)
		hp += int(c.Level-1) * perLevel
	}
	return hp
}

// BaseAC возвращает базовый класс доспеха (с учётом щита, если экипирован и разрешён).
func (c *Character) BaseAC(withShield bool) int {
	rule, ok := ClassRules[c.ClassID]
	if !ok {
		return 10
	}

	ac := rule.BaseAC
	if withShield && c.CanUseShield() {
		ac += 2
	}
	return ac
}

// InitiativeBonus возвращает бонус к инициативе (модификатор Ловкости).
func (c *Character) InitiativeBonus() int {
	return c.Stats.Modifier(c.Stats.Dexterity)
}

// PrimaryStatMod возвращает модификатор основной характеристики класса.
func (c *Character) PrimaryStatMod() int {
	switch c.ClassID {
	case ClassFighter:
		return c.Stats.Modifier(c.Stats.Strength)
	case ClassRanger:
		return c.Stats.Modifier(c.Stats.Dexterity)
	case ClassMystic:
		return c.Stats.Modifier(c.Stats.Willpower)
	default:
		return 0
	}
}

// CanUseShield проверяет, может ли персонаж использовать щит согласно подклассу.
func (c *Character) CanUseShield() bool {
	if c.SubclassID == nil {
		return c.ClassID == ClassFighter
	}
	sub, ok := SubclassRules[*c.SubclassID]
	return ok && sub.CanUseShield
}

// CanWearArmor проверяет допуск персонажа к указанному типу брони.
func (c *Character) CanWearArmor(armor ArmorType) bool {
	if armor == ArmorNone {
		return true
	}
	if c.SubclassID == nil {
		switch c.ClassID {
		case ClassFighter:
			return armor == ArmorLight || armor == ArmorMedium
		case ClassRanger:
			return armor == ArmorLight
		default:
			return false
		}
	}

	sub, ok := SubclassRules[*c.SubclassID]
	if !ok {
		return false
	}
	for _, allowed := range sub.AllowedArmor {
		if allowed == armor {
			return true
		}
	}
	return false
}

// MaxSpellSlots возвращает количество ячеек заклинаний 1-го круга (для Мистика).
func (c *Character) MaxSpellSlots() int {
	rule, ok := ClassRules[c.ClassID]
	if !ok || !rule.IsSpellcaster {
		return 0
	}
	return 1 + int(c.Level) // На 1 уровне: 2 ячейки, далее +1 за уровень
}
