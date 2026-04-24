package service

import (
	"math"
)

type TaxEntry struct {
	AccountHead   string  `json:"account_head"`
	ChargeType    string  `json:"charge_type"` // "On Net Total", "Actual"
	Rate          float64 `json:"rate"`
	Amount        float64 `json:"amount"`
	IsInclusive   bool    `json:"is_inclusive"`
	IsWithholding bool    `json:"is_withholding"`
	Description   string  `json:"description"`
}

type TaxResult struct {
	NetTotal   float64    `json:"net_total"`
	TotalTax   float64    `json:"total_tax"`
	GrandTotal float64    `json:"grand_total"`
	Taxes      []TaxEntry `json:"taxes"`
}

// CalculateTaxes performs generic tax calculations based on document net total.
func CalculateTaxes(netTotal float64, taxes []TaxEntry) TaxResult {
	result := TaxResult{
		NetTotal:   netTotal,
		GrandTotal: netTotal,
		Taxes:      make([]TaxEntry, len(taxes)),
	}

	for i, tax := range taxes {
		var taxAmount float64

		if tax.ChargeType == "Actual" {
			taxAmount = tax.Amount
		} else {
			// Calculation based on Rate
			if tax.IsInclusive {
				// Amount = NetTotal - (NetTotal / (1 + Rate/100))
				taxAmount = netTotal - (netTotal / (1 + (tax.Rate / 100)))
			} else {
				// Amount = NetTotal * (Rate / 100)
				taxAmount = netTotal * (tax.Rate / 100)
			}
		}

		// Rounding to 2 decimal places (Enterprise Standard)
		taxAmount = math.Round(taxAmount*100) / 100

		if tax.IsWithholding {
			// Withholding (PPh) reduces the amount to be paid
			taxAmount = -math.Abs(taxAmount)
			result.GrandTotal += taxAmount
		} else {
			if !tax.IsInclusive {
				// Exclusive tax (PPN) increases the amount to be paid
				result.GrandTotal += taxAmount
			}
			// Inclusive tax is already part of NetTotal/GrandTotal
		}

		result.Taxes[i] = tax
		result.Taxes[i].Amount = taxAmount
		result.TotalTax += taxAmount
	}

	// Final Grand Total rounding
	result.GrandTotal = math.Round(result.GrandTotal*100) / 100
	result.TotalTax = math.Round(result.TotalTax*100) / 100

	return result
}
