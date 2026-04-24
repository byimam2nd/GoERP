package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
)

type PaymentReference struct {
	ReferenceDoctype string  `json:"reference_doctype"`
	ReferenceName    string  `json:"reference_name"`
	AllocatedAmount  float64 `json:"allocated_amount"`
	WriteOffAmount   float64 `json:"write_off_amount"`
	WriteOffAccount  string  `json:"write_off_account"`
}

// ProcessPayment melunasi invoice dan mengupdate status outstanding serta menangani write-off.
func ProcessPayment(tenantID string, paymentDoc map[string]interface{}, references []PaymentReference) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		paymentNo := paymentDoc["name"].(string)

		for _, ref := range references {
			totalToClear := ref.AllocatedAmount + ref.WriteOffAmount
			if totalToClear <= 0 {
				continue
			}

			// 1. Update Outstanding di Invoice
			tableName := "tab" + ref.ReferenceDoctype
			var outstanding float64
			tx.Table(tableName).Where("name = ? AND tenant_id = ?", ref.ReferenceName, tenantID).Select("outstanding_amount").Scan(&outstanding)

			newOutstanding := outstanding - totalToClear
			if newOutstanding < -0.01 { // Allow tiny margin for rounding
				return fmt.Errorf("Overpayment on %s: %s", ref.ReferenceDoctype, ref.ReferenceName)
			}

			updateData := map[string]interface{}{
				"outstanding_amount": newOutstanding,
			}
			if newOutstanding <= 0 {
				updateData["status"] = "Paid"
				updateData["outstanding_amount"] = 0
			}

			if err := tx.Table(tableName).Where("name = ? AND tenant_id = ?", ref.ReferenceName, tenantID).Updates(updateData).Error; err != nil {
				return err
			}

			// 2. Handle Write-off (Jika ada)
			if ref.WriteOffAmount != 0 {
				writeOffEntries := []GLEntry{
					{
						Account:      ref.WriteOffAccount,
						Debit:        ref.WriteOffAmount,
						Credit:       0,
						VoucherType:  "Payment Entry",
						VoucherNo:    paymentNo,
						Remarks:      fmt.Sprintf("Write-off for %s", ref.ReferenceName),
					},
					{
						Account:      "Accounts Receivable", // Target AR Account
						Debit:        0,
						Credit:       ref.WriteOffAmount,
						VoucherType:  "Payment Entry",
						VoucherNo:    paymentNo,
						Remarks:      fmt.Sprintf("Write-off Adjustment for %s", ref.ReferenceName),
					},
				}
				if err := PostGLEntries(tx, writeOffEntries, tenantID); err != nil {
					return err
				}
			}

			// 2. Simpan Relasi Rekonsiliasi (tabPaymentReference)
			refData := map[string]interface{}{
				"tenant_id":         tenantID,
				"parent":            paymentNo,
				"reference_doctype": ref.ReferenceDoctype,
				"reference_name":    ref.ReferenceName,
				"allocated_amount":  ref.AllocatedAmount,
			}
			if err := tx.Table("tabPaymentReference").Create(refData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
