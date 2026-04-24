package service

import (
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type AgingBucket struct {
	Range       string  `json:"range"`
	Amount      float64 `json:"amount"`
	Outstanding float64 `json:"outstanding"`
}

type PartyAging struct {
	PartyName   string        `json:"party_name"`
	Total       float64       `json:"total"`
	Outstanding float64       `json:"outstanding"`
	Buckets     []AgingBucket `json:"buckets"`
}

// GetAgingReport calculates aging for Sales Invoices (Receivable) or Purchase Invoices (Payable)
func GetAgingReport(tenantID, company, partyType string, asOfDate time.Time) ([]PartyAging, error) {
	var results []PartyAging

	tableName := "tabSalesInvoice"
	partyField := "customer"
	if partyType == "Supplier" {
		tableName = "tabPurchaseInvoice"
		partyField = "supplier"
	}

	// 1. Fetch all outstanding invoices
	var invoices []map[string]interface{}
	err := database.DB.Table(tableName).
		Where("tenant_id = ? AND company = ? AND outstanding_amount > 0 AND docstatus = 1 AND posting_date <= ?", tenantID, company, asOfDate).
		Find(&invoices).Error
	
	if err != nil {
		return nil, err
	}

	// Group by Party
	partyInvoices := make(map[string][]map[string]interface{})
	for _, inv := range invoices {
		p := inv[partyField].(string)
		partyInvoices[p] = append(partyInvoices[p], inv)
	}

	// 2. Process each party
	for party, invs := range partyInvoices {
		aging := PartyAging{
			PartyName: party,
			Buckets: []AgingBucket{
				{Range: "0-30 days"},
				{Range: "31-60 days"},
				{Range: "61-90 days"},
				{Range: "Over 90 days"},
			},
		}

		for _, inv := range invs {
			dueDateStr := inv["due_date"].(string)
			dueDate, _ := time.Parse("2006-01-02", dueDateStr)
			outstanding := inv["outstanding_amount"].(float64)

			daysOverdue := int(asOfDate.Sub(dueDate).Hours() / 24)

			if daysOverdue <= 30 {
				aging.Buckets[0].Amount += outstanding
			} else if daysOverdue <= 60 {
				aging.Buckets[1].Amount += outstanding
			} else if daysOverdue <= 90 {
				aging.Buckets[2].Amount += outstanding
			} else {
				aging.Buckets[3].Amount += outstanding
			}

			aging.Outstanding += outstanding
			aging.Total += inv["grand_total"].(float64)
		}
		results = append(results, aging)
	}

	return results, nil
}
