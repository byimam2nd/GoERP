package service

import (
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type LedgerEntryDetail struct {
	PostingDate time.Time `json:"posting_date"`
	VoucherType string    `json:"voucher_type"`
	VoucherNo   string    `json:"voucher_no"`
	Against     string    `json:"against"`
	Debit       float64   `json:"debit"`
	Credit      float64   `json:"credit"`
	Balance     float64   `json:"balance"`
	Remarks     string    `json:"remarks"`
}

type AccountExplanation struct {
	AccountName    string              `json:"account_name"`
	OpeningBalance float64             `json:"opening_balance"`
	Entries        []LedgerEntryDetail `json:"entries"`
	ClosingBalance float64             `json:"closing_balance"`
}

// ExplainAccountBalance provide a detailed audit trail of an account balance
func ExplainAccountBalance(tenantID, company, account string, start, end time.Time) (*AccountExplanation, error) {
	// 1. Calculate Opening Balance
	var opening float64
	database.DB.Table("tabGLEntry").
		Select("SUM(debit - credit)").
		Where("tenant_id = ? AND company = ? AND account = ? AND posting_date < ? AND docstatus = 1", tenantID, company, account, start).
		Scan(&opening)

	// 2. Fetch Entries within period
	var entries []LedgerEntryDetail
	database.DB.Table("tabGLEntry").
		Select("posting_date, voucher_type, voucher_no, against, debit, credit, remarks").
		Where("tenant_id = ? AND company = ? AND account = ? AND posting_date BETWEEN ? AND ? AND docstatus = 1", tenantID, company, account, start, end).
		Order("posting_date ASC, creation ASC").
		Scan(&entries)

	// 3. Compute running balance
	runningBalance := opening
	for i := range entries {
		runningBalance += (entries[i].Debit - entries[i].Credit)
		entries[i].Balance = runningBalance
	}

	return &AccountExplanation{
		AccountName:    account,
		OpeningBalance: opening,
		Entries:        entries,
		ClosingBalance: runningBalance,
	}, nil
}
