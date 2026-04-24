package selling

import (
	"context"
	"fmt"
	"math"
	"github.com/goerp/goerp/apps/erp/accounting/service"
	"time"
)

type SalesInvoiceLedger struct{}

func (g *SalesInvoiceLedger) GenerateGL(ctx context.Context, doc map[string]interface{}) ([]service.GLEntry, error) {
	grandTotal, _ := doc["grand_total"].(float64)
	netTotal, _ := doc["net_total"].(float64)
	customer, _ := doc["customer"].(string)
	voucherNo, _ := doc["name"].(string)
	postingDateStr, _ := doc["posting_date"].(string)
	
	postingDate, _ := time.Parse("2006-01-02", postingDateStr)
	if postingDate.IsZero() {
		postingDate = time.Now()
	}

	var entries []service.GLEntry

	// 1. Debit Accounts Receivable (Grand Total)
	entries = append(entries, service.GLEntry{
		Account:     "Accounts Receivable",
		Debit:       grandTotal,
		Credit:      0,
		Against:     "Sales Income",
		Remarks:     "Sales Invoice Billing",
		VoucherType: "Sales Invoice",
		VoucherNo:   voucherNo,
		PostingDate: postingDate,
	})

	// 2. Credit Sales Income (Net Total)
	entries = append(entries, service.GLEntry{
		Account:     "Sales Income",
		Debit:       0,
		Credit:      netTotal,
		Against:     customer,
		Remarks:     "Sales Invoice Revenue",
		VoucherType: "Sales Invoice",
		VoucherNo:   voucherNo,
		PostingDate: postingDate,
	})

	// 3. Process Taxes from child table
	if taxData, ok := doc["taxes"].([]interface{}); ok {
		for _, t := range taxData {
			taxMap, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			accountHead, _ := taxMap["account_head"].(string)
			taxAmount, _ := taxMap["amount"].(float64)
			isWithholding, _ := taxMap["is_withholding"].(bool)

			if taxAmount == 0 {
				continue
			}

			if isWithholding {
				// Withholding Tax (PPh) - Usually Debit for Sales (Tax paid by customer on our behalf)
				// OR Credit if we are tracking it as a liability to pay later. 
				// In Indonesia, PPh 23 Sales means customer deducts, so we record as Prepaid Tax (Asset).
				entries = append(entries, service.GLEntry{
					Account:     accountHead,
					Debit:       math.Abs(taxAmount),
					Credit:      0,
					Against:     customer,
					Remarks:     fmt.Sprintf("Withholding Tax: %s", accountHead),
					VoucherType: "Sales Invoice",
					VoucherNo:   voucherNo,
					PostingDate: postingDate,
				})
			} else {
				// Regular Tax (PPN) - Credit (Liability)
				entries = append(entries, service.GLEntry{
					Account:     accountHead,
					Debit:       0,
					Credit:      taxAmount,
					Against:     customer,
					Remarks:     fmt.Sprintf("Tax: %s", accountHead),
					VoucherType: "Sales Invoice",
					VoucherNo:   voucherNo,
					PostingDate: postingDate,
				})
			}
		}
	}

	return entries, nil
}
