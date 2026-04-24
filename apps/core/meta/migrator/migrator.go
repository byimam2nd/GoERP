package migrator

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/types"
	"go.uber.org/zap"
)

func MigrateDocType(dt *types.DocType) error {
	tableName := fmt.Sprintf("tab%s", dt.Name)
	
	// Check if table exists
	if !database.DB.Migrator().HasTable(tableName) {
		// Create basic table with ID and metadata fields
		createSQL := fmt.Sprintf(`CREATE TABLE "%s" (
			name VARCHAR(255) PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL,
			creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_by VARCHAR(255),
			owner VARCHAR(255),
			docstatus INT DEFAULT 0
		)`, tableName)
		
		if err := database.DB.Exec(createSQL).Error; err != nil {
			return err
		}
		
		// Create an index on tenant_id for performance
		indexSQL := fmt.Sprintf(`CREATE INDEX "idx_%s_tenant" ON "%s" (tenant_id)`, dt.Name, tableName)
		database.DB.Exec(indexSQL)

		logger.Log.Info("Created table with tenant isolation", zap.String("table", tableName))
	}

	// Add fields as columns
	for _, field := range dt.Fields {
		columnType := getColumnType(field.FieldType)
		if columnType == "" {
			continue
		}

		if !database.DB.Migrator().HasColumn(tableName, field.Name) {
			alterSQL := fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s`, tableName, field.Name, columnType)
			if field.Required {
				alterSQL += " NOT NULL DEFAULT ''" // Simplified for string-based
			}

			if err := database.DB.Exec(alterSQL).Error; err != nil {
				logger.Log.Error("Failed to add column", 
					zap.String("table", tableName), 
					zap.String("column", field.Name), 
					zap.Error(err))
				return err
			}

			if field.Unique {
				indexName := fmt.Sprintf("ux_%s_%s", dt.Name, field.Name)
				uniqueIndexSQL := fmt.Sprintf(`CREATE UNIQUE INDEX "%s" ON "%s" ("%s", "tenant_id")`, indexName, tableName, field.Name)
				database.DB.Exec(uniqueIndexSQL)
			}

			logger.Log.Info("Added column", zap.String("table", tableName), zap.String("column", field.Name))
		}

		// 3. Handle Indexing
		if field.SearchIndex {
			indexName := fmt.Sprintf("idx_%s_%s", dt.Name, field.Name)
			createIndexSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "%s" ON "%s" ("%s")`, indexName, tableName, field.Name)
			database.DB.Exec(createIndexSQL)
		}
	}

	// 4. Special Optimization for Ledgers (Composite Indexes)
	if dt.Name == "GLEntry" || dt.Name == "StockLedgerEntry" {
		// Optimization for financial reports: (account, posting_date)
		database.DB.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "idx_%s_fin_report" ON "%s" (account, posting_date)`, dt.Name, tableName))
		// Optimization for tenant isolation + date: (tenant_id, posting_date)
		database.DB.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "idx_%s_tenant_date" ON "%s" (tenant_id, posting_date)`, dt.Name, tableName))
	}

	return nil
}

func getColumnType(ft types.FieldType) string {
	switch ft {
	case types.Data, types.Select, types.Link:
		return "VARCHAR(255)"
	case types.Int:
		return "BIGINT"
	case types.Float, types.Currency:
		return "DECIMAL(18, 6)"
	case types.Date:
		return "DATE"
	case types.DateTime:
		return "TIMESTAMP"
	case types.Check:
		return "BOOLEAN"
	case types.Text:
		return "TEXT"
	default:
		return ""
	}
}
