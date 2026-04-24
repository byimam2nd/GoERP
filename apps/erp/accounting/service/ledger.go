package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"math"
	"time"
)

type GLEntry struct {
	Account        string
	PostingDate    time.Time
	Currency       string
	ExchangeRate   float64
	Debit          float64
	Credit         float64
	DebitForeign   float64
	CreditForeign  float64
	Against        string
	CostCenter     string 
	Project        string
	VoucherType    string
	VoucherNo      string
}

func PostGLEntries(db *gorm.DB, entries []GLEntry, tenantID string) error {
	if db == nil {
		db = database.DB
	}

	// 0. Numerical Guard & Multi-currency normalization
	for i := range entries {
		entries[i].Debit = math.Round(entries[i].Debit*100) / 100
		entries[i].Credit = math.Round(entries[i].Credit*100) / 100
		
		if entries[i].ExchangeRate == 0 {
			entries[i].ExchangeRate = 1.0
		}
		
		// If foreign values are not provided but currency is foreign, calculate them
		if entries[i].Currency != "" && entries[i].DebitForeign == 0 && entries[i].CreditForeign == 0 {
			entries[i].DebitForeign = entries[i].Debit / entries[i].ExchangeRate
			entries[i].CreditForeign = entries[i].Credit / entries[i].ExchangeRate
		}

		if entries[i].PostingDate.IsZero() {
			entries[i].PostingDate = time.Now()
		}
	}

	// 1. Validate Balance (Debit must equal Credit)
	var totalDebit, totalCredit float64
	for _, e := range entries {
		totalDebit += e.Debit
		totalCredit += e.Credit
	}

	if math.Abs(totalDebit-totalCredit) > 0.0001 {
		return fmt.Errorf("GL Unbalanced: Debit (%.2f) != Credit (%.2f)", totalDebit, totalCredit)
	}

	// 2. Period Locking Validation (The Fiscal Guard)
	for _, e := range entries {
		if e.VoucherNo == "" {
			return fmt.Errorf("Explainability Error: Every GL Entry must have a VoucherNo")
		}

		var fiscalYear struct {
			IsClosed bool
		}
		// Cek apakah tanggal transaksi berada di periode yang sudah dikunci
		err := db.Table("tabFiscalYear").
			Select("is_closed").
			Where("tenant_id = ? AND ? BETWEEN year_start_date AND year_end_date", tenantID, e.PostingDate).
			Scan(&fiscalYear).Error
		
		if err == nil && fiscalYear.IsClosed {
			return fmt.Errorf("Fiscal Guard: Period for date %s is CLOSED. Cannot post entries.", e.PostingDate.Format("2006-01-02"))
		}
	}

	tx := db.Begin()

	for _, e := range entries {
		glName := fmt.Sprintf("GL-%d-%s", time.Now().UnixNano(), e.Account)
		data := map[string]interface{}{
			"name": glName, "tenant_id": tenantID, "account": e.Account,
			"debit": e.Debit, "credit": e.Credit,
			"currency": e.Currency, "exchange_rate": e.ExchangeRate,
			"debit_foreign": e.DebitForeign, "credit_foreign": e.CreditForeign,
			"against": e.Against, "cost_center": e.CostCenter, "project": e.Project,
			"voucher_type": e.VoucherType, "voucher_no": e.VoucherNo,
			"posting_date": e.PostingDate, "creation": time.Now(), "modified": time.Now(), "docstatus": 1,
		}

		if err := tx.Table("tabGLEntry").Create(data).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 3. Incremental Balance Update (Global Account Balance)
		balanceChange := e.Debit - e.Credit
		
		// Postgres specific UPSERT. For SQLite it would be different, but we'll stick to standard PG as per GEMINI.md
		upsertSQL := `INSERT INTO "tabAccountBalance" (account, tenant_id, balance, last_updated) 
					  VALUES (?, ?, ?, ?)
					  ON CONFLICT (account, tenant_id) 
					  DO UPDATE SET balance = "tabAccountBalance".balance + EXCLUDED.balance, 
					                last_updated = EXCLUDED.last_updated`
		
		if err := tx.Exec(upsertSQL, e.Account, tenantID, balanceChange, time.Now()).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update balance for %s: %v", e.Account, err)
		}
	}

func ReverseGLEntries(db *gorm.DB, voucherType string, voucherNo string, tenantID string) error {
	if db == nil {
		db = database.DB
	}

	// 1. Check for Active Dependencies before reversing
	var activeDependencies []struct {
		ChildDT   string
		ChildName string
	}
	db.Table("tabDocLink").
		Where("tenant_id = ? AND parent_name = ? AND docstatus = 1", tenantID, voucherNo).
		Scan(&activeDependencies)

	if len(activeDependencies) > 0 {
		return fmt.Errorf("Integrity Error: Cannot reverse %s %s. Active dependencies found: %v", 
			voucherType, voucherNo, activeDependencies)
	}

	var originalEntries []GLEntry
	err := db.Table("tabGLEntry").
		Where("tenant_id = ? AND voucher_type = ? AND voucher_no = ? AND docstatus = 1", tenantID, voucherType, voucherNo).
		Find(&originalEntries).Error
	if err != nil {
		return err
	}

	if len(originalEntries) == 0 {
		return nil 
	}

	tx := db.Begin()
	
	// Mark original as cancelled (docstatus = 2)
	if err := tx.Table("tabGLEntry").
		Where("tenant_id = ? AND voucher_type = ? AND voucher_no = ?", tenantID, voucherType, voucherNo).
		Update("docstatus", 2).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create and post reverse entries
	var reverseEntries []GLEntry
	for _, e := range originalEntries {
		reverseEntries = append(reverseEntries, GLEntry{
			Account:      e.Account,
			PostingDate:  time.Now(),
			Debit:        e.Credit,   
			Credit:       e.Debit,
			Against:      e.Against,
			VoucherType:  e.VoucherType,
			VoucherNo:    e.VoucherNo,
			Remarks:      fmt.Sprintf("Cancellation Reversal: %s", e.VoucherNo),
		})
	}

	if err := PostGLEntries(tx, reverseEntries, tenantID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
