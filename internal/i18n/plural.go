package i18n

// PluralRule defines a function that returns the plural form index for a count.
type PluralRule func(count int) int

// Standard plural rules for common languages.
var pluralRules = map[string]PluralRule{
	// Chinese, Japanese, Korean, Vietnamese - no plural forms
	"zh": pluralRuleSingle,
	"zh-CN": pluralRuleSingle,
	"zh-TW": pluralRuleSingle,
	"ja": pluralRuleSingle,
	"ko": pluralRuleSingle,
	"vi": pluralRuleSingle,

	// English, German, Dutch, Swedish, etc. - singular (1) vs plural (n != 1)
	"en": pluralRuleGermanic,
	"en-US": pluralRuleGermanic,
	"en-GB": pluralRuleGermanic,
	"de": pluralRuleGermanic,
	"nl": pluralRuleGermanic,
	"sv": pluralRuleGermanic,

	// French, Portuguese - singular (0, 1) vs plural (n > 1)
	"fr": pluralRuleFrench,
	"pt": pluralRuleFrench,
	"pt-BR": pluralRuleFrench,

	// Polish - complex plural forms
	"pl": pluralRulePolish,

	// Russian, Ukrainian - complex plural forms
	"ru": pluralRuleSlavic,
	"uk": pluralRuleSlavic,
}

// pluralRuleSingle - no plural distinction (Chinese, Japanese, etc.)
func pluralRuleSingle(count int) int {
	return 0
}

// pluralRuleGermanic - English-style: one (1) vs other
func pluralRuleGermanic(count int) int {
	if count == 1 {
		return 0 // "one" form
	}
	return 1 // "other" form
}

// pluralRuleFrench - French-style: one (0, 1) vs other
func pluralRuleFrench(count int) int {
	if count == 0 || count == 1 {
		return 0 // "one" form
	}
	return 1 // "other" form
}

// pluralRulePolish - Polish complex rules
func pluralRulePolish(count int) int {
	if count == 1 {
		return 0 // "one"
	}
	mod10 := count % 10
	mod100 := count % 100
	if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
		return 1 // "few"
	}
	return 2 // "other"
}

// pluralRuleSlavic - Russian-style complex rules
func pluralRuleSlavic(count int) int {
	mod10 := count % 10
	mod100 := count % 100

	if mod10 == 1 && mod100 != 11 {
		return 0 // "one"
	}
	if mod10 >= 2 && mod10 <= 4 && !(mod100 >= 12 && mod100 <= 14) {
		return 1 // "few"
	}
	return 2 // "other"
}

// GetPluralRule returns the plural rule for a language.
// Falls back to Germanic (English-style) if not found.
func GetPluralRule(lang string) PluralRule {
	if rule, ok := pluralRules[lang]; ok {
		return rule
	}

	// Try base language (e.g., "en" for "en-US")
	if idx := findDot(lang); idx > 0 {
		base := lang[:idx]
		if rule, ok := pluralRules[base]; ok {
			return rule
		}
	}

	// Default to Germanic
	return pluralRuleGermanic
}

// findDot finds the first dot in a string.
func findDot(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' || s[i] == '-' {
			return i
		}
	}
	return -1
}
