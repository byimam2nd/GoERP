package service

import (
	"github.com/goerp/goerp/apps/core/domain"
	"gorm.io/gorm"
	"time"
)

// AccountingBridge adalah adapter yang menghubungkan modul Accounting ke kernel domain.
type AccountingBridge struct{}

func (b *AccountingBridge) PostJournal(tx *gorm.DB, tenantID string, entries []domain.AccountingEntry) error {
	// Konversi domain.AccountingEntry ke service.GLEntry
	var glEntries []GLEntry
	for _, e := range entries {
		glEntries = append(glEntries, GLEntry{
			Account:     e.Account,
			Debit:       e.Debit,
			Credit:      e.Credit,
			PostingDate: e.PostingDate,
			VoucherType: e.VoucherType,
			VoucherNo:   e.VoucherNo,
			CostCenter:  e.CostCenter,
			Remarks:     e.Remarks,
		})
	}
	return PostGLEntries(tx, glEntries, tenantID)
}

func (b *AccountingBridge) CheckBudget(tenantID string, costCenter string, account string, amount float64, date time.Time) (bool, string, error) {
	res, err := CheckBudget(tenantID, costCenter, account, amount, date)
	if err != nil {
		return false, "", err
	}
	// Return Stop if IsExceeded
	stop := res.IsExceeded && res.Action == "Stop"
	return stop, res.Message, nil
}

func (b *AccountingBridge) GetCompanyDefaultAccount(tenantID string, company string, fieldName string) (string, error) {
	// Di V2, ini seharusnya pindah ke BaseModule bridge, 
	// tapi untuk sekarang kita ambil dari service yang ada.
	// (Note: butuh refactor GetCompanyDefaultAccount agar bisa diakses tanpa circular dependency)
	return "", nil // Placeholder untuk demo pemisahan
}
