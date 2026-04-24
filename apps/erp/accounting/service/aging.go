package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type AgingBucket struct {
	Range      string  `json:"range"`
	Amount     float64 `json:"amount"`
}

type AgingResult struct {
	Party      string        `json:"party"`
	Total      float64       `json:"total"`
	Buckets    []AgingBucket `json:"buckets"`
}

// GetAgingReport menghitung outstanding AR/AP berdasarkan umur piutang/hutang.
func GetAgingReport(tenantID string, partyType string, partyName string) (AgingResult, error) {
	// partyType: "Customer" (AR) or "Supplier" (AP)
	tableName := "tabSalesInvoice"
	if partyType == "Supplier" {
		tableName = "tabPurchaseInvoice"
	}

	var results []struct {
		Name              string
		OutstandingAmount float64
		PostingDate       time.Time
	}

	err := database.DB.Table(tableName).
		Select("name, outstanding_amount, posting_date").
		Where("tenant_id = ? AND outstanding_amount > 0 AND status != 'Paid'", tenantID).
		Scan(&results).Error

	if err != nil {
		return AgingResult{}, err
	}

	now := time.Now()
	aging := AgingResult{
		Party:   partyName,
		Buckets: []AgingBucket{
			{Range: "0-30 days", Amount: 0},
			{Range: "31-60 days", Amount: 0},
			{Range: "61-90 days", Amount: 0},
			{Range: "90+ days", Amount: 0},
		},
	}

	for _, r := range results {
		days := int(now.Sub(r.PostingDate).Hours() / 24)
		aging.Total += r.OutstandingAmount

		if days <= 30 {
			aging.Buckets[0].Amount += r.OutstandingAmount
		} else if days <= 60 {
			aging.Buckets[1].Amount += r.OutstandingAmount
		} else if days <= 90 {
			aging.Buckets[2].Amount += r.OutstandingAmount
		} else {
			aging.Buckets[3].Amount += r.OutstandingAmount
		}
	}

	return aging, nil
}
