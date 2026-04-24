package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
)

type FinancialNarrative struct {
	Account   string `json:"account"`
	Balance   float64 `json:"balance"`
	Trace     []TraceNode `json:"trace"`
}

type TraceNode struct {
	VoucherNo       string    `json:"voucher_no"`
	VoucherType     string    `json:"voucher_type"`
	Amount          float64   `json:"amount"`
	SourceDoc       string    `json:"source_doc,omitempty"`
	SystemTimestamp time.Time `json:"system_timestamp"` // Forensic: Real insertion time
}

// GetForensicBalance mengambil saldo akun pada titik waktu sistem tertentu (as-of timestamp).
func GetForensicBalance(tenantID string, accountName string, asOfSystem time.Time) (float64, error) {
	var balance float64
	// Query ini mengabaikan transaksi yang di-input backdated SETELAH asOfSystem
	err := database.DB.Table("tabGLEntry").
		Where("tenant_id = ? AND account = ? AND creation <= ?", tenantID, accountName, asOfSystem).
		Select("SUM(debit - credit)").Scan(&balance).Error
	return balance, err
}

// GetNarrativeTrace membangun silsilah angka finansial untuk auditor.
func GetNarrativeTrace(tenantID string, accountName string) (FinancialNarrative, error) {
	var entries []struct {
		VoucherNo   string
		VoucherType string
		Amount      float64
	}

	// 1. Ambil Jurnal
	database.DB.Table("tabGLEntry").
		Where("tenant_id = ? AND account = ? AND docstatus = 1", tenantID, accountName).
		Select("voucher_no, voucher_type, (debit - credit) as amount").
		Scan(&entries)

	narrative := FinancialNarrative{Account: accountName}
	for _, e := range entries {
		// 2. Untuk setiap jurnal, cari dokumen sumbernya (Lineage)
		var source string
		database.DB.Table("tabDocLink").
			Where("tenant_id = ? AND child_name = ?", tenantID, e.VoucherNo).
			Select("parent_name").Limit(1).Scan(&source)

		narrative.Trace = append(narrative.Trace, TraceNode{
			VoucherNo:   e.VoucherNo,
			VoucherType: e.VoucherType,
			Amount:      e.Amount,
			SourceDoc:   source,
		})
		narrative.Balance += e.Amount
	}

	return narrative, nil
}
