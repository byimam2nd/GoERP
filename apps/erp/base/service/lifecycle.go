package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
)

type DocumentLink struct {
	ParentDocType string
	ParentName    string
	ChildDocType  string
	ChildName     string
	ItemCode      string
	Qty           float64
}

// UpdateLifecycleStatus updates pending quantities and status across document chains
func UpdateLifecycleStatus(link DocumentLink, tenantID string) error {
	tx := database.DB.Begin()

	// Example logic for Sales Order -> Sales Invoice
	// We need to update 'invoiced_qty' in Sales Order Item
	if link.ParentDocType == "Sales Order" && link.ChildDocType == "Sales Invoice" {
		updateItemSQL := `UPDATE "tabSalesOrderItem" 
						 SET invoiced_qty = invoiced_qty + ? 
						 WHERE parent = ? AND item_code = ? AND tenant_id = ?`
		
		if err := tx.Exec(updateItemSQL, link.Qty, link.ParentName, link.ItemCode, tenantID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Check if fully invoiced to update Parent Status
		var pending float64
		checkSQL := `SELECT SUM(qty - invoiced_qty) FROM "tabSalesOrderItem" WHERE parent = ? AND tenant_id = ?`
		tx.Raw(checkSQL, link.ParentName, tenantID).Scan(&pending)

		status := "Partially Invoiced"
		if pending <= 0 {
			status = "Fully Invoiced"
		}

		if err := tx.Table("tabSalesOrder").
			Where("name = ? AND tenant_id = ?", link.ParentName, tenantID).
			Update("status", status).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// MapDocument performs "Make Sales Invoice" logic from Sales Order
func MapDocument(fromDocType, fromName, toDocType string, tenantID string) (map[string]interface{}, error) {
	// This is the "Engine Mapper" that automatically fetches data from SO to SI
	// including only the REMAINING quantity.
	
	// 1. Fetch source document and items
	var sourceDoc map[string]interface{}
	database.DB.Table("tab"+fromDocType).Where("name = ? AND tenant_id = ?", fromName, tenantID).First(&sourceDoc)
	
	var items []map[string]interface{}
	itemTable := "tab" + fromDocType + "Item"
	database.DB.Table(itemTable).
		Where("parent = ? AND tenant_id = ? AND (qty - invoiced_qty) > 0", fromName, tenantID).
		Find(&items)

	if len(items) == 0 {
		return nil, fmt.Errorf("no pending items to map from %s", fromName)
	}

	// 2. Build target document
	target := make(map[string]interface{})
	target["customer"] = sourceDoc["customer"]
	target["company"] = sourceDoc["company"]
	target["from_voucher_no"] = fromName
	
	var targetItems []map[string]interface{}
	for _, itm := range items {
		qty := itm["qty"].(float64) - itm["invoiced_qty"].(float64)
		targetItems = append(targetItems, map[string]interface{}{
			"item_code": itm["item_code"],
			"qty":       qty,
			"rate":      itm["rate"],
			"amount":    qty * itm["rate"].(float64),
			"so_detail": itm["name"],
		})
	}
	target["items"] = targetItems

	return target, nil
}
