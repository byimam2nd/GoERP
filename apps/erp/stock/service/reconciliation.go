package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
)

type ReconciliationReport struct {
	Account      string  `json:"account"`
	GLBalance    float64 `json:"gl_balance"`
	StockValue   float64 `json:"stock_value"`
	Difference   float64 `json:"difference"`
}

// CheckStockGLConsistency membandingkan nilai inventaris di GL vs Gudang (Bin).
func CheckStockGLConsistency(tenantID string, stockAccount string) (ReconciliationReport, error) {
	// 1. Get GL Balance for the Stock Account
	var glBalance float64
	database.DB.Table("tabAccountBalance").
		Where("tenant_id = ? AND account = ?", tenantID, stockAccount).
		Select("balance").Scan(&glBalance)

	// 2. Get Stock Value from all Warehouses linked to this account
	// In GoERP, we assume a mapping or warehouse linked to the account
	var stockValue float64
	query := `
		SELECT SUM(valuation_rate * actual_qty) 
		FROM tabBin 
		WHERE tenant_id = ? 
		AND warehouse IN (SELECT name FROM tabWarehouse WHERE account = ?)
	`
	database.DB.Raw(query, tenantID, stockAccount).Scan(&stockValue)

	return ReconciliationReport{
		Account:    stockAccount,
		GLBalance:  glBalance,
		StockValue: stockValue,
		Difference: glBalance - stockValue,
	}, nil
}
