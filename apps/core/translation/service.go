package translation

import (
	"github.com/goerp/goerp/apps/core/database"
)

var cache = make(map[string]map[string]string) // language -> key -> value

// Translate returns the translated string for a key in a specific language
func Translate(key string, lang string, tenantID string) string {
	// 1. Check Cache
	if lCache, ok := cache[lang]; ok {
		if val, ok := lCache[key]; ok {
			return val
		}
	}

	// 2. Fallback to Database
	var translation map[string]interface{}
	err := database.DB.Table("tabTranslation").
		Where("source_text = ? AND language = ? AND tenant_id = ?", key, lang, tenantID).
		First(&translation).Error

	if err == nil {
		val := translation["translated_text"].(string)
		// Update Cache (simplified)
		return val
	}

	return key // Return original if no translation found
}
