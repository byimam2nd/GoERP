package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

type StockInput struct {
	ItemCode         string
	Warehouse        string
	Qty              float64
	Rate             float64 
	VoucherType      string
	VoucherNo        string
	PostingDate      time.Time
	TenantID         string
	InventoryAccount string 
	DeferredAccount  string 
}

func UpdateStock(input StockInput) error {
	if input.PostingDate.IsZero() {
		input.PostingDate = time.Now()
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Check for backdated transaction
		var latestSLE map[string]interface{}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND tenant_id = ? AND docstatus = 1", input.ItemCode, input.Warehouse, input.TenantID).
			Order("posting_date DESC, creation DESC").
			Limit(1).
			Find(&latestSLE)

		isBackdated := false
		if latestSLE != nil && latestSLE["posting_date"] != nil {
			if input.PostingDate.Before(latestSLE["posting_date"].(time.Time)) {
				isBackdated = true
			}
		}

		// 2. Insert Current SLE
		var prevSLE map[string]interface{}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND tenant_id = ? AND posting_date <= ? AND docstatus = 1", input.ItemCode, input.Warehouse, input.TenantID, input.PostingDate).
			Order("posting_date DESC, creation DESC").
			Limit(1).
			Find(&prevSLE)

		qtyAfter := input.Qty
		prevValue := 0.0
		if prevSLE != nil && prevSLE["qty_after_transaction"] != nil {
			qtyAfter += prevSLE["qty_after_transaction"].(float64)
			prevValue = prevSLE["stock_value"].(float64)
		}

		valuationRate := input.Rate
		if input.Qty < 0 {
			valuationRate = calculateFIFORate(tx, input.ItemCode, input.Warehouse, input.TenantID, -input.Qty)
		}

		stockValueChange := input.Qty * valuationRate
		newStockValue := prevValue + stockValueChange

		sle := map[string]interface{}{
			"name": fmt.Sprintf("SLE-%d", time.Now().UnixNano()), "tenant_id": input.TenantID,
			"item_code": input.ItemCode, "warehouse": input.Warehouse, "actual_qty": input.Qty,
			"qty_after_transaction": qtyAfter, "valuation_rate": valuationRate, "stock_value": newStockValue,
			"voucher_type": input.VoucherType, "voucher_no": input.VoucherNo,
			"posting_date": input.PostingDate, "creation": time.Now(), "modified": time.Now(), "docstatus": 1,
		}
		if err := tx.Table("tabStockLedgerEntry").Create(sle).Error; err != nil { return err }

		// 3. Trigger Reposting if backdated
		if isBackdated {
			// In production, use Asynq worker
			go RepostItemValuation(input.ItemCode, input.Warehouse, input.TenantID, input.PostingDate)
		} else {
			// Regular update for Bin if not backdated
			tx.Table("tabBin").Where("item_code = ? AND warehouse = ? AND tenant_id = ?", input.ItemCode, input.Warehouse, input.TenantID).
				Updates(map[string]interface{}{"actual_qty": qtyAfter, "stock_value": newStockValue, "valuation_rate": valuationRate})
		}

		return nil
	})
}

func calculateFIFORate(tx *gorm.DB, item, warehouse, tenantID string, qtyToIssue float64) float64 {
	var incomingEntries []map[string]interface{}
	tx.Table("tabStockLedgerEntry").
		Where("item_code = ? AND warehouse = ? AND tenant_id = ? AND actual_qty > 0 AND docstatus = 1", item, warehouse, tenantID).
		Order("posting_date ASC, creation ASC").
		Find(&incomingEntries)

	if len(incomingEntries) > 0 {
		return incomingEntries[0]["valuation_rate"].(float64)
	}
	return 0
}

func RepostItemValuation(item, warehouse, tenantID string, startingFrom time.Time) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var sles []map[string]interface{}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND tenant_id = ? AND posting_date >= ? AND docstatus = 1", 
				item, warehouse, tenantID, startingFrom).
			Order("posting_date ASC, creation ASC, name ASC").
			Find(&sles)

		var prevSLE map[string]interface{}
		tx.Table("tabStockLedgerEntry").
			Where("item_code = ? AND warehouse = ? AND tenant_id = ? AND posting_date < ? AND docstatus = 1", 
				item, warehouse, tenantID, startingFrom).
			Order("posting_date DESC, creation DESC").
			Limit(1).
			Find(&prevSLE)

		qtyAfter := 0.0
		stockValue := 0.0
		if prevSLE != nil && prevSLE["qty_after_transaction"] != nil {
			qtyAfter = prevSLE["qty_after_transaction"].(float64)
			stockValue = prevSLE["stock_value"].(float64)
		}

		for _, sle := range sles {
			actualQty := sle["actual_qty"].(float64)
			valuationRate := sle["valuation_rate"].(float64)
			if actualQty < 0 { valuationRate = calculateFIFORate(tx, item, warehouse, tenantID, -actualQty) }
			qtyAfter += actualQty
			stockValue += (actualQty * valuationRate)
			tx.Table("tabStockLedgerEntry").Where("name = ?", sle["name"]).Updates(map[string]interface{}{
				"qty_after_transaction": qtyAfter, "valuation_rate": valuationRate, "stock_value": stockValue, "modified": time.Now(),
			})
		}
		return tx.Table("tabBin").Where("item_code = ? AND warehouse = ? AND tenant_id = ?", item, warehouse, tenantID).
			Updates(map[string]interface{}{"actual_qty": qtyAfter, "stock_value": stockValue, "modified": time.Now()}).Error
	})
}
