package meta

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type LedgerEngine struct {
	generators map[string]LedgerGenerator
}

var DefaultLedgerEngine = &LedgerEngine{
	generators: make(map[string]LedgerGenerator),
}

func (e *LedgerEngine) RegisterGenerator(doctype string, g LedgerGenerator) {
	e.generators[doctype] = g
}

func (e *LedgerEngine) Process(ctx context.Context, doctype string, doc map[string]interface{}, isCancel bool) error {
	gen, ok := e.generators[doctype]
	if !ok {
		return nil // No accounting/stock effect for this DocType
	}

	tenantID := doc["tenant_id"].(string)
	docName := doc["name"].(string)

	// 1. Generate Entries
	glEntries, err := gen.GenerateGL(ctx, doc)
	if err != nil {
		return err
	}
	stockEntries, err := gen.GenerateStock(ctx, doc)
	if err != nil {
		return err
	}

	// 2. Wrap in Transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// A. Process GL
		for _, gle := range glEntries {
			if isCancel {
				// Reverse Debit/Credit
				gle.Debit, gle.Credit = -gle.Debit, -gle.Credit
			}
			
			data := map[string]interface{}{
				"name":         fmt.Sprintf("GL-%d-%s", time.Now().UnixNano(), gle.Account),
				"tenant_id":    tenantID,
				"posting_date": time.Now(),
				"account":      gle.Account,
				"debit":        gle.Debit,
				"credit":       gle.Credit,
				"against":      gle.Against,
				"voucher_type": doctype,
				"voucher_no":   docName,
				"remarks":      gle.Remarks,
				"creation":     time.Now(),
				"modified":     time.Now(),
				"docstatus":    1,
			}
			if err := tx.Table("tabGLEntry").Create(data).Error; err != nil {
				return err
			}
		}

		// B. Process Stock (Integration with Bin)
		for _, se := range stockEntries {
			qty := se.Qty
			if isCancel {
				qty = -qty
			}

			// Update Bin
			var bin map[string]interface{}
			err := tx.Table("tabBin").
				Where("item_code = ? AND warehouse = ? AND tenant_id = ?", se.ItemCode, se.Warehouse, tenantID).
				First(&bin).Error

			if err == nil {
				currentQty := bin["actual_qty"].(float64)
				tx.Table("tabBin").Where("name = ?", bin["name"]).Update("actual_qty", currentQty+qty)
			} else {
				// Create new bin if doesn't exist
				newBin := map[string]interface{}{
					"name":       fmt.Sprintf("BIN-%d", time.Now().UnixNano()),
					"tenant_id":  tenantID,
					"item_code":  se.ItemCode,
					"warehouse":  se.Warehouse,
					"actual_qty": qty,
					"creation":   time.Now(),
					"modified":   time.Now(),
				}
				tx.Table("tabBin").Create(newBin)
			}

			// Create SLE
			sle := map[string]interface{}{
				"name":         fmt.Sprintf("SLE-%d", time.Now().UnixNano()),
				"tenant_id":    tenantID,
				"item_code":    se.ItemCode,
				"warehouse":    se.Warehouse,
				"qty":          qty,
				"voucher_type": doctype,
				"voucher_no":   docName,
				"posting_date": time.Now(),
				"creation":     time.Now(),
				"modified":     time.Now(),
				"docstatus":    1,
			}
			tx.Table("tabStockLedgerEntry").Create(sle)
		}

		logger.Log.Info("Ledger Engine: Transaction processed", 
			zap.String("doctype", doctype), 
			zap.String("name", docName), 
			zap.Bool("is_cancel", isCancel))

		return nil
	})
}
