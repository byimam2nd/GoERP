package handlers

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/job"
	"time"
)

func RegisterMaintenanceHandlers() {
	job.RegisterHandler("PartitionMaintenance", HandlePartitionMaintenance)
}

func HandlePartitionMaintenance(ctx context.Context, tenantID string, payload map[string]interface{}) error {
	// 1. Tentukan target bulan (bulan depan)
	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)
	year := nextMonth.Year()
	month := int(nextMonth.Month())

	// 2. Daftar tabel yang harus dipartisi
	tables := []string{"tabGLEntry", "tabStockLedgerEntry", "tabActivityLog"}

	for _, table := range tables {
		partitionName := fmt.Sprintf("%s_y%dm%02d", table, year, month)
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

		// 3. Eksekusi DDL untuk membuat partisi
		query := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
			FOR VALUES FROM ('%s') TO ('%s')
		`, partitionName, table, startDate, endDate)

		if err := database.DB.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to create partition %s: %v", partitionName, err)
		}
	}

	return nil
}
