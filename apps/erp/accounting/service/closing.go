package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

// CloseFiscalYear melakukan penutupan tahun buku:
// 1. Menghitung Laba/Rugi dari akun Income & Expense.
// 2. Memindahkan saldo ke Retained Earnings.
// 3. Mengunci periode tersebut.
func CloseFiscalYear(ctx context.Context, tenantID string, fiscalYearName string, retainedEarningsAccount string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil info Fiscal Year
		var fy struct {
			YearStartDate time.Time
			YearEndDate   time.Time
			IsClosed      bool
		}
		if err := tx.Table("tabFiscalYear").Where("name = ? AND tenant_id = ?", fiscalYearName, tenantID).First(&fy).Error; err != nil {
			return fmt.Errorf("fiscal year not found: %v", err)
		}
		if fy.IsClosed {
			return fmt.Errorf("fiscal year is already closed")
		}

		// 2. Hitung Total Revenue & Expense
		var plBalance float64
		// Asumsi: Akun Income memiliki root_type 'Asset' atau 'Liability' tertentu, 
		// di GoERP kita filter berdasarkan report_type 'Profit and Loss'
		query := `
			SELECT SUM(debit - credit) 
			FROM tabGLEntry 
			WHERE tenant_id = ? 
			AND posting_date BETWEEN ? AND ?
			AND account IN (SELECT name FROM tabAccount WHERE report_type = 'Profit and Loss')
		`
		tx.Raw(query, tenantID, fy.YearStartDate, fy.YearEndDate).Scan(&plBalance)

		if plBalance == 0 {
			return fmt.Errorf("no balance to close for this period")
		}

		// 3. Buat Closing Entry (Pindah ke Retained Earnings)
		// Jika plBalance negatif (Credit > Debit) -> Profit
		// Jika plBalance positif (Debit > Credit) -> Loss
		entries := []GLEntry{
			{
				Account:      retainedEarningsAccount,
				Debit:        0,
				Credit:       -plBalance, // Jika profit, Credit Retained Earnings
				VoucherType:  "Period Closing",
				VoucherNo:    fmt.Sprintf("CLOSE-%s", fiscalYearName),
				PostingDate:  fy.YearEndDate,
				Remarks:      "Closing of Profit/Loss to Retained Earnings",
			},
			{
				Account:      "P&L Summary", // Akun sementara untuk kliring
				Debit:        plBalance,
				Credit:       0,
				VoucherType:  "Period Closing",
				VoucherNo:    fmt.Sprintf("CLOSE-%s", fiscalYearName),
				PostingDate:  fy.YearEndDate,
				Remarks:      "P&L Summary Clearing",
			},
		}

		if err := PostGLEntries(tx, entries, tenantID); err != nil {
			return err
		}

		// 4. Kunci Fiscal Year
		return tx.Table("tabFiscalYear").
			Where("name = ? AND tenant_id = ?", fiscalYearName, tenantID).
			Update("is_closed", true).Error
	})
}
