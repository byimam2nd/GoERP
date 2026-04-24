package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
)

type BOMItem struct {
	ItemCode string  `json:"item_code"`
	Qty      float64 `json:"qty"`
	Rate     float64 `json:"rate"`
}

// GetBOMCost menghitung biaya produksi berdasarkan daftar bahan baku.
func GetBOMCost(tenantID string, bomName string) (float64, error) {
	var totalCost float64
	var items []BOMItem

	// Ambil semua bahan baku di dalam BOM
	query := `
		SELECT item_code, qty, 
		       (SELECT valuation_rate FROM tabBin WHERE item_code = tabBOMItem.item_code AND tenant_id = ? LIMIT 1) as rate
		FROM tabBOMItem
		WHERE parent = ? AND tenant_id = ?
	`
	err := database.DB.Raw(query, tenantID, bomName, tenantID).Scan(&items).Error
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		totalCost += item.Qty * item.Rate
	}

	return totalCost, nil
}
