package service

import (
	"github.com/goerp/goerp/apps/core/database"
)

type ReconciliationResult struct {
	Account      string  `json:"account"`
	GLBalance    float64 `json:"gl_balance"`
	StockValue   float64 `json:"stock_value"`
	Difference   float64 `json:"difference"`
	IsReconciled bool    `json:"is_reconciled"`
}

// CheckStockGLReconciliation compares GL Account Balances with Actual Stock Value
func CheckStockGLReconciliation(tenantID, company string) ([]ReconciliationResult, error) {
	// 1. Get Accounts of type 'Stock'
	var stockAccounts []string
	database.DB.Table("tabAccount").
		Where("tenant_id = ? AND company = ? AND account_type = 'Stock'", tenantID, company).
		Pluck("name", &stockAccounts)

	var results []ReconciliationResult

	for _, acc := range stockAccounts {
		// 2. Get GL Balance
		var glBalance float64
		database.DB.Table("tabAccountBalance").
			Where("account = ? AND tenant_id = ?", acc, tenantID).
			Select("balance").Scan(&glBalance)

		// 3. Get Stock Value (Sum of Valuation Rate * Actual Qty from Bins)
		// We assume Bins are linked to Accounts via Warehouse or directly
		// For this MVP, we sum all Bins for items that use this Stock Account
		var stockValue float64
		query := `SELECT SUM(actual_qty * valuation_rate) 
				  FROM "tabBin" 
				  WHERE tenant_id = ? AND warehouse IN (
					  SELECT name FROM "tabWarehouse" WHERE account = ?
				  )`
		
		database.DB.Raw(query, tenantID, acc).Scan(&stockValue)

		diff := glBalance - stockValue
		results = append(results, ReconciliationResult{
			Account:      acc,
			GLBalance:    glBalance,
			StockValue:   stockValue,
			Difference:   diff,
			IsReconciled: (diff < 0.01 && diff > -0.01),
		})
	}

	return results, nil
}
