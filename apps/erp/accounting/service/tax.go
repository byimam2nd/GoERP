package service

import (
	"math"
)

type TaxType string

const (
	OnNetTotal     TaxType = "On Net Total"
	OnPreviousRow  TaxType = "On Previous Row Amount"
)

type TaxRow struct {
	AccountHead   string  `json:"account_head"`
	Description   string  `json:"description"`
	Rate          float64 `json:"rate"`
	TaxAmount     float64 `json:"tax_amount"`
	TotalAmount   float64 `json:"total_amount"`
	IncludedInPrice bool  `json:"included_in_price"`
}

type TaxResult struct {
	NetTotal   float64  `json:"net_total"`
	TaxTotal   float64  `json:"tax_total"`
	GrandTotal float64  `json:"grand_total"`
	Taxes      []TaxRow `json:"taxes"`
}

// CalculateTaxes performs the core tax logic
func CalculateTaxes(baseAmount float64, taxRows []TaxRow) TaxResult {
	result := TaxResult{
		NetTotal: baseAmount,
		Taxes:    make([]TaxRow, len(taxRows)),
	}

	runningTotal := baseAmount
	var totalTax float64

	for i, row := range taxRows {
		var amount float64
		
		if row.IncludedInPrice {
			// Back-calculate from inclusive price: Tax = Total - (Total / (1 + Rate/100))
			amount = runningTotal - (runningTotal / (1 + (row.Rate / 100)))
			// For inclusive, net total actually decreases
			result.NetTotal -= amount
		} else {
			// Standard exclusive calculation
			amount = runningTotal * (row.Rate / 100)
		}

		// Precision Rounding (2 decimal places for financial)
		amount = math.Round(amount*100) / 100
		
		result.Taxes[i] = row
		result.Taxes[i].TaxAmount = amount
		
		if !row.IncludedInPrice {
			totalTax += amount
		}
	}

	result.TaxTotal = totalTax
	result.GrandTotal = result.NetTotal + totalTax
	
	return result
}
