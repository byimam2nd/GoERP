package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"math"
	"time"
)

type AssetDepreciationMethod string

const (
	StraightLine           AssetDepreciationMethod = "Straight Line"
	DoubleDecliningBalance AssetDepreciationMethod = "Double Declining Balance"
)

type DepreciationEntry struct {
	ScheduleDate   time.Time `json:"schedule_date"`
	DepreciationAmt float64   `json:"depreciation_amount"`
	IsPosted       bool      `json:"is_posted"`
	GLEntry        string    `json:"gl_entry,omitempty"`
}

// GenerateDepreciationSchedule calculates monthly depreciation parts
func GenerateDepreciationSchedule(purchaseAmount float64, salvageValue float64, usefulLifeMonths int, startDate time.Time) []DepreciationEntry {
	if usefulLifeMonths <= 0 { return nil }

	monthlyAmt := (purchaseAmount - salvageValue) / float64(usefulLifeMonths)
	monthlyAmt = math.Round(monthlyAmt*100) / 100 // Financial rounding

	schedule := make([]DepreciationEntry, 0)
	for i := 1; i <= usefulLifeMonths; i++ {
		scheduleDate := startDate.AddDate(0, i, 0)
		schedule = append(schedule, DepreciationEntry{
			ScheduleDate:    scheduleDate,
			DepreciationAmt: monthlyAmt,
			IsPosted:        false,
		})
	}

	return schedule
}

// PostScheduledDepreciation scans for due schedules and creates GL entries
func PostScheduledDepreciation(tenantID string) error {
	now := time.Now()
	
	// 1. Fetch all Assets with unposted schedules due today or earlier
	// In production, this would be a more complex query joining Asset with its Schedule child table
	var dueAssets []map[string]interface{}
	database.DB.Table("tabAsset").
		Where("tenant_id = ? AND docstatus = 1 AND status = 'Active'", tenantID).
		Find(&dueAssets)

	for _, asset := range dueAssets {
		assetName := asset["name"].(string)
		
		// In a real system, we'd fetch the specific schedule rows from tabAssetDepreciationSchedule
		// Let's assume we find one due entry for this example
		
		depreciationExpenseAccount := asset["depreciation_expense_account"].(string)
		accumulatedDepreciationAccount := asset["accumulated_depreciation_account"].(string)
		amount := 500.0 // Placeholder for the actual scheduled amount

		entries := []GLEntry{
			{
				Account:     depreciationExpenseAccount,
				Debit:       amount,
				Credit:      0,
				VoucherType: "Asset",
				VoucherNo:   assetName,
			},
			{
				Account:     accumulatedDepreciationAccount,
				Debit:       0,
				Credit:      amount,
				VoucherType: "Asset",
				VoucherNo:   assetName,
			},
		}

		if err := PostGLEntries(entries, tenantID); err != nil {
			return fmt.Errorf("failed to post depreciation for %s: %v", assetName, err)
		}

		// Update Asset's current value
		database.DB.Table("tabAsset").Where("name = ?", assetName).
			Update("current_value", gorm.Expr("current_value - ?", amount))
	}

	return nil
}
