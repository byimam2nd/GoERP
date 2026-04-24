package multitenant

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/auth"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

func ProvisionTenant(tenantName, domain, adminEmail, adminPassword string) error {
	tenantID := fmt.Sprintf("T-%d", time.Now().UnixNano())

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create Tenant Record
		tenantData := map[string]interface{}{
			"name":        tenantID,
			"tenant_id":   "system", // The tenant registry itself belongs to system
			"tenant_name": tenantName,
			"domain":      domain,
			"is_active":   true,
			"creation":    time.Now(),
			"modified":    time.Now(),
		}
		if err := tx.Table("tabTenant").Create(tenantData).Error; err != nil {
			return err
		}

		// 2. Create Default Administrator Role for the new tenant
		roleData := map[string]interface{}{
			"name":      "Administrator",
			"tenant_id": tenantID,
			"role_name": "Administrator",
			"creation":  time.Now(),
			"modified":  time.Now(),
		}
		if err := tx.Table("tabRole").Create(roleData).Error; err != nil {
			return err
		}

		// 3. Create initial Administrator User
		hashedPassword, _ := auth.HashPassword(adminPassword)
		userData := map[string]interface{}{
			"name":      "admin",
			"tenant_id": tenantID,
			"username":  "admin",
			"email":     adminEmail,
			"password":  hashedPassword,
			"full_name": "Administrator",
			"creation":  time.Now(),
			"modified":  time.Now(),
		}
		if err := tx.Table("tabUser").Create(userData).Error; err != nil {
			return err
		}

		// 4. Seed basic permissions for the Administrator role
		coreDocTypes := []string{"User", "Role", "RolePermission", "Tenant", "DocType", "Company"}
		for _, dt := range coreDocTypes {
			permData := map[string]interface{}{
				"name":            fmt.Sprintf("PERM-%s-%s", tenantID, dt),
				"tenant_id":       tenantID,
				"role":            "Administrator",
				"parent_doctype":  dt,
				"read":            true,
				"write":           true,
				"can_create":      true,
				"delete":          true,
				"creation":        time.Now(),
				"modified":        time.Now(),
			}
			if err := tx.Table("tabRolePermission").Create(permData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
