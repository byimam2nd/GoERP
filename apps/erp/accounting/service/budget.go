package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type BudgetCheckResult struct {
	IsExceeded bool
	Action     string // "Stop" or "Warn"
	Message    string
}

// CheckBudget memeriksa apakah sebuah transaksi melampaui sisa anggaran.
func CheckBudget(tenantID string, costCenter string, account string, amount float64, postingDate time.Time) (BudgetCheckResult, error) {
	// 1. Cari Budget yang aktif untuk kombinasi Cost Center + Account + Fiscal Year
	var budget struct {
		BudgetAmount     float64
		ActionIfExceeded string
	}
	
	err := database.DB.Table("tabBudget").
		Where("tenant_id = ? AND cost_center = ? AND account = ?", tenantID, costCenter, account).
		// Simplifikasi: mencari fiscal year berdasarkan tanggal posting
		Where("? BETWEEN (SELECT year_start_date FROM tabFiscalYear WHERE name = tabBudget.fiscal_year) AND (SELECT year_end_date FROM tabFiscalYear WHERE name = tabBudget.fiscal_year)", postingDate).
		Scan(&budget).Error

	if err != nil || budget.BudgetAmount == 0 {
		return BudgetCheckResult{IsExceeded: false}, nil // Tidak ada budget yang diatur
	}

	// 2. Hitung Realisasi Saat Ini (Actual Expenses) dari GL
	var actualSpent float64
	database.DB.Table("tabGLEntry").
		Where("tenant_id = ? AND cost_center = ? AND account = ? AND docstatus = 1", tenantID, costCenter, account).
		Select("SUM(debit - credit)").Scan(&actualSpent)

	// 3. Evaluasi Dampak Transaksi Baru
	totalProjected := actualSpent + amount
	if totalProjected > budget.BudgetAmount {
		return BudgetCheckResult{
			IsExceeded: true,
			Action:     budget.ActionIfExceeded,
			Message:    fmt.Sprintf("Budget Exceeded for %s. Limit: %.2f, Actual: %.2f, This Transaction: %.2f", account, budget.BudgetAmount, actualSpent, amount),
		}, nil
	}

	return BudgetCheckResult{IsExceeded: false}, nil
}
