package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type ClosingRequest struct {
	FiscalYear      string    `json:"fiscal_year"`
	ClosingDate     time.Time `json:"closing_date"`
	RetainedEarnings string    `json:"retained_earnings_account"`
	Company         string    `json:"company"`
	TenantID        string    `json:"tenant_id"`
}

// ClosePeriod zeroes out P&L accounts and transfers balance to Retained Earnings
func ClosePeriod(req ClosingRequest) error {
	tx := database.DB.Begin()

	// 1. Fetch P&L Accounts and their Balances
	var plBalances []map[string]interface{}
	// Join Account with AccountBalance to get only Income/Expense
	query := `SELECT b.account, b.balance 
			  FROM "tabAccountBalance" b
			  JOIN "tabAccount" a ON a.name = b.account AND a.tenant_id = b.tenant_id
			  WHERE a.root_type IN ('Income', 'Expense') 
			  AND a.company = ? AND a.tenant_id = ?`
	
	if err := tx.Raw(query, req.Company, req.TenantID).Scan(&plBalances).Error; err != nil {
		tx.Rollback()
		return err
	}

	var glEntries []GLEntry
	var totalPLBalance float64

	// 2. Create Reversing Entries for each P&L Account
	voucherNo := fmt.Sprintf("CLOSE-%s", req.FiscalYear)
	for _, b := range plBalances {
		acc := b["account"].(string)
		balance := b["balance"].(float64)
		if balance == 0 { continue }

		entry := GLEntry{
			Account:     acc,
			PostingDate: req.ClosingDate,
			VoucherType: "PeriodClosingVoucher",
			VoucherNo:   voucherNo,
		}

		if balance > 0 {
			// Current is Debit (usually Expense), so we Credit it
			entry.Credit = balance
		} else {
			// Current is Credit (usually Income), so we Debit it
			entry.Debit = -balance
		}
		
		glEntries = append(glEntries, entry)
		totalPLBalance += balance
	}

	// 3. Post to Retained Earnings (The balancing entry)
	if totalPLBalance != 0 {
		reEntry := GLEntry{
			Account:     req.RetainedEarnings,
			PostingDate: req.ClosingDate,
			VoucherType: "PeriodClosingVoucher",
			VoucherNo:   voucherNo,
		}
		if totalPLBalance > 0 {
			// Net Loss: Debit Retained Earnings
			reEntry.Debit = totalPLBalance
		} else {
			// Net Profit: Credit Retained Earnings
			reEntry.Credit = -totalPLBalance
		}
		glEntries = append(glEntries, reEntry)
	}

	// 4. Use existing PostGLEntries within this transaction
	if err := PostGLEntries(tx, glEntries, req.TenantID); err != nil {
		tx.Rollback()
		return err
	}

	// 5. Mark Fiscal Year as Closed
	if err := tx.Table("tabFiscalYear").
		Where("year = ? AND tenant_id = ?", req.FiscalYear, req.TenantID).
		Update("is_closed", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
