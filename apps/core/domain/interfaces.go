package domain

import (
	"context"
	"gorm.io/gorm"
	"time"
)

// AccountingEntry adalah representasi generik entri jurnal untuk dikirim antar modul.
type AccountingEntry struct {
	Account      string
	Debit        float64
	Credit       float64
	PostingDate  time.Time
	VoucherType  string
	VoucherNo    string
	CostCenter   string
	Remarks      string
}

// IAccountingModule mendefinisikan apa yang modul lain boleh minta dari Accounting.
type IAccountingModule interface {
	PostJournal(tx *gorm.DB, tenantID string, entries []AccountingEntry) error
	CheckBudget(tenantID string, costCenter string, account string, amount float64, date time.Time) (bool, string, error)
	GetCompanyDefaultAccount(tenantID string, company string, fieldName string) (string, error)
}

// Registry menyimpan implementasi konkret dari setiap domain.
var (
	Accounting IAccountingModule
)

func RegisterAccounting(impl IAccountingModule) {
	Accounting = impl
}
