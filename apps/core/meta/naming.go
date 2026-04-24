package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"strings"
	"time"
)

func GetNextName(doctype string, tenantID string) (string, error) {
	var series map[string]interface{}
	err := database.DB.Table("tabNamingSeries").
		Where("ref_doctype = ? AND tenant_id = ?", doctype, tenantID).
		First(&series).Error

	if err != nil {
		return fmt.Sprintf("%s-%d", doctype, time.Now().UnixNano()), nil
	}

	prefix := series["prefix"].(string)
	currentValue := int(series["current_value"].(int64))
	padding := int(series["padding"].(int64))

	now := time.Now()
	prefix = strings.ReplaceAll(prefix, ".YYYY.", now.Format("2006"))
	prefix = strings.ReplaceAll(prefix, ".MM.", now.Format("01"))
	prefix = strings.ReplaceAll(prefix, ".DD.", now.Format("02"))

	newValue := currentValue + 1
	
	if err := database.DB.Table("tabNamingSeries").
		Where("name = ? AND tenant_id = ?", series["name"], tenantID).
		Update("current_value", newValue).Error; err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%dd", padding)
	name := fmt.Sprintf("%s%s", prefix, fmt.Sprintf(format, newValue))

	return name, nil
}

func GetAmendedName(originalName string) string {
	// If already has suffix -X, increment it
	parts := strings.Split(originalName, "-")
	lastPart := parts[len(parts)-1]
	
	var version int
	_, err := fmt.Sscanf(lastPart, "%d", &version)
	
	if err == nil {
		// It's already a versioned name like INV-0001-1
		newVersion := version + 1
		parts[len(parts)-1] = fmt.Sprintf("%d", newVersion)
		return strings.Join(parts, "-")
	}

	// First time amendment: INV-0001 -> INV-0001-1
	return fmt.Sprintf("%s-1", originalName)
}
