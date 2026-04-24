package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"math"
	"time"
)

type RevaluationRequest struct {
	Company         string    `json:"company"`
	TenantID        string    `json:"tenant_id"`
	Date            time.Time `json:"date"`
	GainLossAccount string    `json:"gain_loss_account"`
}

// GetExchangeRate returns the exchange rate for a given currency to base currency on a specific date
func GetExchangeRate(fromCurrency, toCurrency, tenantID string, date time.Time) (float64, error) {
	var record struct {
		ExchangeRate float64
	}
	err := database.DB.Table("tabCurrencyExchange").
		Select("exchange_rate").
		Where("tenant_id = ? AND from_currency = ? AND to_currency = ? AND date <= ?", tenantID, fromCurrency, toCurrency, date).
		Order("date DESC").
		Limit(1).
		Scan(&record).Error
	
	if err != nil || record.ExchangeRate == 0 {
		return 1.0, fmt.Errorf("exchange rate not found for %s to %s on %v", fromCurrency, toCurrency, date)
	}
	return record.ExchangeRate, nil
}

// RunCurrencyRevaluation identifies accounts that need revaluation and posts gain/loss entries
func RunCurrencyRevaluation(req RevaluationRequest) error {
	tx := database.DB.Begin()

	// 1. Identify base currency
	var baseCurrency string
	tx.Table("tabCurrency").Where("tenant_id = ? AND is_base = 1", req.TenantID).Select("currency_name").Scan(&baseCurrency)
	if baseCurrency == "" {
		return fmt.Errorf("base currency not defined for tenant %s", req.TenantID)
	}

	// 2. Fetch all non-base currency accounts with balance
	var accounts []map[string]interface{}
	query := `SELECT a.name, a.account_currency, b.balance as base_balance 
			  FROM "tabAccount" a
			  JOIN "tabAccountBalance" b ON a.name = b.account AND a.tenant_id = b.tenant_id
			  WHERE a.tenant_id = ? AND a.company = ? 
			  AND a.account_currency IS NOT NULL AND a.account_currency != ?`
	
	if err := tx.Raw(query, req.TenantID, req.Company, baseCurrency).Scan(&accounts).Error; err != nil {
		tx.Rollback()
		return err
	}

	var glEntries []GLEntry
	voucherNo := fmt.Sprintf("REVAL-%d", time.Now().Unix())

	for _, acc := range accounts {
		name := acc["name"].(string)
		currency := acc["account_currency"].(string)
		baseBalance := acc["base_balance"].(float64)

		if baseBalance == 0 { continue }

		// 3. Get latest exchange rate
		rate, err := GetExchangeRate(currency, baseCurrency, req.TenantID, req.Date)
		if err != nil {
			fmt.Printf("Warning: skipping revaluation for %s: %v\n", name, err)
			continue
		}

		// Logic:
		// We assume 'baseBalance' in GoERP is ALWAYS stored in base currency.
		// However, for foreign accounts, we should ideally track both:
		// 'foreign_balance' (USD) and 'base_balance' (IDR).
		// For simplicity in this engine, we assume there's a way to know the original foreign balance.
		// In a production ERP, GLEntry would have 'debit_foreign' and 'credit_foreign'.
		
		// For this MVP, let's assume we have a table 'tabAccountForeignBalance' 
		// or we query GL entries for that account to sum original foreign values.
		
		var foreignBalance float64
		tx.Table("tabGLEntry").
			Select("SUM(debit_foreign - credit_foreign)").
			Where("account = ? AND tenant_id = ?", name, req.TenantID).
			Scan(&foreignBalance)
		
		if foreignBalance == 0 { continue }

		expectedBaseBalance := foreignBalance * rate
		adjustment := expectedBaseBalance - baseBalance

		if math.Abs(adjustment) < 0.01 { continue }

		// 4. Create GL Entries for adjustment
		// Entry for the Foreign Account
		accEntry := GLEntry{
			Account:     name,
			PostingDate: req.Date,
			VoucherType: "CurrencyRevaluation",
			VoucherNo:   voucherNo,
		}
		// Entry for Unrealized Gain/Loss
		glEntry := GLEntry{
			Account:     req.GainLossAccount,
			PostingDate: req.Date,
			VoucherType: "CurrencyRevaluation",
			VoucherNo:   voucherNo,
		}

		if adjustment > 0 {
			// Increase in asset/Decrease in liability value
			accEntry.Debit = adjustment
			glEntry.Credit = adjustment // Gain
		} else {
			// Decrease in asset/Increase in liability value
			accEntry.Credit = -adjustment
			glEntry.Debit = -adjustment // Loss
		}

		glEntries = append(glEntries, accEntry, glEntry)
	}

	if len(glEntries) > 0 {
		if err := PostGLEntries(tx, glEntries, req.TenantID); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
