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
