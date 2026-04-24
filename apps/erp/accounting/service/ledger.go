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

	// 2. Period Locking Validation
	for _, e := range entries {
		var fiscalYear struct {
			IsClosed bool
		}
		err := db.Table("tabFiscalYear").
			Select("is_closed").
			Where("tenant_id = ? AND ? BETWEEN year_start_date AND year_end_date", tenantID, e.PostingDate).
			Scan(&fiscalYear).Error
		
		if err != nil {
			return fmt.Errorf("failed to check fiscal year for date %v: %v", e.PostingDate, err)
		}
		if fiscalYear.IsClosed {
			return fmt.Errorf("cannot post to closed fiscal year for date %v", e.PostingDate)
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

	return tx.Commit().Error
}
