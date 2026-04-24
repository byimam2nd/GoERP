package selling

import (
	"context"
	"github.com/goerp/goerp/apps/core/meta"
)

type SalesInvoiceLedger struct{}

func (g *SalesInvoiceLedger) GenerateGL(ctx context.Context, doc map[string]interface{}) ([]meta.GLEntry, error) {
	total, _ := doc["total"].(float64)
	customer, _ := doc["customer"].(string)

	return []meta.GLEntry{
		{
			Account: "Accounts Receivable",
			Debit:   total,
			Credit:  0,
			Against: customer,
			Remarks: "Sales Invoice Billing",
		},
		{
			Account: "Sales Income",
			Debit:   0,
			Credit:  total,
			Against: customer,
			Remarks: "Sales Invoice Revenue",
		},
	}, nil
}

func (g *SalesInvoiceLedger) GenerateStock(ctx context.Context, doc map[string]interface{}) ([]meta.StockEntry, error) {
	// In a real system, this would iterate over child table 'items'
	// For this demo, let's assume one item from the form
	return []meta.StockEntry{
		{
			ItemCode:  "DEMO-ITEM",
			Warehouse: "Main Warehouse",
			Qty:       -1, // Outgoing
			Rate:      100,
		},
	}, nil
}
