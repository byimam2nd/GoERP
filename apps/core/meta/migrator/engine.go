package migrator

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/types"
	"gorm.io/gorm"
)

// ExecutePlan runs the migration steps within a transaction with an advisory lock.
func ExecutePlan(ctx context.Context, plan MigrationPlan) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Get Advisory Lock (Postgres Specific) to prevent concurrent migrations on the same DocType
		var lockAcquired bool
		// Using a hash of the DocType name as the lock ID
		lockQuery := fmt.Sprintf("SELECT pg_try_advisory_xact_lock(42, hashtext('%s'))", plan.DocType)
		if err := tx.Raw(lockQuery).Scan(&lockAcquired).Error; err != nil {
			return fmt.Errorf("failed to acquire migration lock: %v", err)
		}
		if !lockAcquired {
			return fmt.Errorf("migration for DocType %s is already in progress by another process", plan.DocType)
		}

		// 2. Execute each step
		for _, step := range plan.Steps {
			if err := executeStep(tx, step); err != nil {
				return err
			}
		}

		// 3. Record migration history
		history := map[string]interface{}{
			"parent":   plan.DocType,
			"version":  plan.Version,
			"creation": gorm.Expr("NOW()"),
			"details":  fmt.Sprintf("Executed %d steps", len(plan.Steps)),
		}
		
		// Ensure tabDocTypeHistory exists or handle error
		if err := tx.Table("tabDocTypeHistory").Create(history).Error; err != nil {
			// If table doesn't exist, we might want to skip or auto-create it
			// For now, let's just log it or handle it gracefully
			fmt.Printf("Warning: Could not record history to tabDocTypeHistory: %v\n", err)
		}

		return nil
	})
}

func executeStep(tx *gorm.DB, step MigrationStep) error {
	sqlType := getSQLType(step.Field.FieldType)
	tableName := "tab" + step.TableName

	switch step.Type {
	case AddColumn:
		// Check if column exists first to be idempotent
		var exists bool
		checkQuery := `SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = ? AND column_name = ?
		)`
		tx.Raw(checkQuery, tableName, step.Field.Name).Scan(&exists)
		if exists {
			return nil // Already added
		}

		query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, step.Field.Name, sqlType)
		return tx.Exec(query).Error

	case AlterColumn:
		// Postgres specific ALTER COLUMN with USING for casting
		query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", 
			tableName, step.Field.Name, sqlType, step.Field.Name, sqlType)
		return tx.Exec(query).Error

	case CreateIndex:
		indexName := fmt.Sprintf("idx_%s_%s", tableName, step.Field.Name)
		query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, step.Field.Name)
		return tx.Exec(query).Error
	}

	return nil
}

func getSQLType(ft types.FieldType) string {
	switch ft {
	case types.Currency, types.Float:
		return "DECIMAL(18,4)"
	case types.Int:
		return "INTEGER"
	case types.Check:
		return "BOOLEAN"
	case types.Date:
		return "DATE"
	case types.DateTime:
		return "TIMESTAMP"
	case types.Text:
		return "TEXT"
	default:
		return "VARCHAR(255)"
	}
}
