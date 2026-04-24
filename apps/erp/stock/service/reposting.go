package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"time"
)

type RepostRequest struct {
	ItemCode  string    `json:"item_code"`
	Warehouse string    `json:"warehouse"`
	FromDate  time.Time `json:"from_date"`
	TenantID  string    `json:"tenant_id"`
}

// RepostStockLedger recalculates SLEs for a specific item-warehouse from a given date
func RepostStockLedger(ctx context.Context, req RepostRequest) error {
	log := logger.Log.With(
		zap.String("item", req.ItemCode),
		zap.String("warehouse", req.Warehouse),
		zap.String("tenant", req.TenantID),
	)

	log.Info("Starting backdated reposting")

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Get the starting balance before FromDate
		var prevSLE struct {
			BalanceQty    float64
			ValuationRate float64
		}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND posting_date < ? AND tenant_id = ? AND docstatus = 1",
				req.ItemCode, req.Warehouse, req.FromDate, req.TenantID).
			Order("posting_date DESC, creation DESC").
			Limit(1).
			Scan(&prevSLE)

		currentBalance := prevSLE.BalanceQty
		currentValuation := prevSLE.ValuationRate

		// 2. Fetch all SLEs from FromDate onwards
		var entries []map[string]interface{}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND posting_date >= ? AND tenant_id = ? AND docstatus = 1",
				req.ItemCode, req.Warehouse, req.FromDate, req.TenantID).
			Order("posting_date ASC, creation ASC").
			Find(&entries)

		// 3. Re-calculate each entry
		for _, entry := range entries {
			actualQty := entry["actual_qty"].(float64)
			
			// Simple Valuation Logic (Moving Average for brevity in this example)
			// In production, this would involve the complex FIFO queue logic
			if actualQty > 0 {
				incomingRate := entry["incoming_rate"].(float64)
				totalValue := (currentBalance * currentValuation) + (actualQty * incomingRate)
				currentBalance += actualQty
				if currentBalance > 0 {
					currentValuation = totalValue / currentBalance
				}
			} else {
				currentBalance += actualQty // actualQty is negative for outgoing
			}

			// 4. Update the Entry with correct Balance and Valuation
			tx.Table("tabStockLedgerEntry").
				Where("name = ?", entry["name"]).
				Updates(map[string]interface{}{
					"balance_qty":    currentBalance,
					"valuation_rate": currentValuation,
					"modified":       time.Now(),
				})
		}

		log.Info("Reposting completed", zap.Int("entries_updated", len(entries)))
		return nil
	})
}
