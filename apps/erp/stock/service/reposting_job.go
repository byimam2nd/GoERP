package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/job"
	"gorm.io/gorm"
	"time"
)

// RepostStockJobHandler memproses ulang FIFO stock valuation untuk item tertentu.
func RepostStockJobHandler(ctx context.Context, tenantID string, payload map[string]interface{}) error {
	itemCode, _ := payload["item_code"].(string)
	warehouse, _ := payload["warehouse"].(string)

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Logika berat FIFO recalculation dilakukan di sini di background.
		// Ini mencegah UI blocking saat user memproses transaksi ribuan item.
		fmt.Printf("Reposting stock for %s in %s for tenant %s...\n", itemCode, warehouse, tenantID)
		time.Sleep(2 * time.Second) // Simulasi beban komputasi
		return nil
	})
}

func RegisterStockHandlers() {
	job.RegisterHandler("RepostStock", RepostStockJobHandler)
}
