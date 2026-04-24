package service

import (
	"context"
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"time"
)

type RevaluationResult struct {
	Account      string
	Currency     string
	OldBalance   float64
	NewBalance   float64
	GainLoss     float64
}

// RunCurrencyRevaluation menghitung dan memposting Unrealized Gain/Loss untuk akun Valas.
func RunCurrencyRevaluation(ctx context.Context, tenantID string, revalDate time.Time, marketRates map[string]float64) ([]RevaluationResult, error) {
	var results []RevaluationResult

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil semua akun yang memiliki currency bukan Base Currency (IDR)
		var accounts []struct {
			Name            string
			AccountCurrency string
		}
		tx.Table("tabAccount").Where("tenant_id = ? AND account_currency != 'IDR'", tenantID).Scan(&accounts)

		for _, acc := range accounts {
			newRate, ok := marketRates[acc.AccountCurrency]
			if !ok {
				continue
			}

			// 2. Ambil saldo valas saat ini (DebitForeign - CreditForeign)
			var foreignBalance float64
			var baseBalance float64
			tx.Table("tabGLEntry").
				Where("tenant_id = ? AND account = ? AND posting_date <= ?", tenantID, acc.Name, revalDate).
				Select("SUM(debit_foreign - credit_foreign) as foreign_bal, SUM(debit - credit) as base_bal").
				Row().Scan(&foreignBalance, &baseBalance)

			if foreignBalance == 0 {
				continue
			}

			// 3. Hitung nilai seharusnya dalam Base Currency (IDR)
			expectedBaseBalance := foreignBalance * newRate
			gainLoss := expectedBaseBalance - baseBalance

			if gainLoss == 0 {
				continue
			}

			// 4. Post Unrealized Gain/Loss ke GL
			gainLossAccount := "Unrealized Exchange Gain/Loss"
			entries := []GLEntry{
				{
					Account:      acc.Name,
					Debit:        gainLoss,
					Credit:       0,
					VoucherType:  "Currency Revaluation",
					VoucherNo:    fmt.Sprintf("REVAL-%s-%s", acc.Name, revalDate.Format("200601")),
					PostingDate:  revalDate,
					Remarks:      fmt.Sprintf("Revaluation at rate %.2f", newRate),
				},
				{
					Account:      gainLossAccount,
					Debit:        0,
					Credit:       gainLoss,
					VoucherType:  "Currency Revaluation",
					VoucherNo:    fmt.Sprintf("REVAL-%s-%s", acc.Name, revalDate.Format("200601")),
					PostingDate:  revalDate,
					Remarks:      "Unrealized Gain/Loss Offset",
				},
			}

			if err := PostGLEntries(tx, entries, tenantID); err != nil {
				return err
			}

			results = append(results, RevaluationResult{
				Account:    acc.Name,
				Currency:   acc.AccountCurrency,
				OldBalance: baseBalance,
				NewBalance: expectedBaseBalance,
				GainLoss:   gainLoss,
			})
		}
		return nil
	})

	return results, err
}
