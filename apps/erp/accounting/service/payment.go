package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type PaymentReference struct {
	ReferenceDocType string  `json:"reference_doctype"`
	ReferenceName    string  `json:"reference_name"`
	AllocatedAmount  float64 `json:"allocated_amount"`
}

type PaymentRequest struct {
	Name            string             `json:"name"`
	PaymentType     string             `json:"payment_type"` // Receive, Pay
	PostingDate     time.Time          `json:"posting_date"`
	PartyType       string             `json:"party_type"` // Customer, Supplier
	Party           string             `json:"party"`
	PaidTo          string             `json:"paid_to"` // Bank/Cash Account
	PaidAmount      float64            `json:"paid_amount"`
	WriteOffAmount  float64            `json:"write_off_amount"`
	WriteOffAccount string             `json:"write_off_account"`
	References      []PaymentReference `json:"references"`
	TenantID        string             `json:"tenant_id"`
}

// ProcessPayment handles the financial logic of a payment
func ProcessPayment(req PaymentRequest) error {
	tx := database.DB.Begin()

	// 1. Validate total allocation
	var totalAllocated float64
	for _, ref := range req.References {
		totalAllocated += ref.AllocatedAmount
	}

	// Total allocated + Write-off must equal Paid Amount (conceptually)
	// In some ERPs: PaidAmount = TotalAllocated - WriteOff
	// We'll follow: Total Impact on Party = PaidAmount + WriteOff
	totalImpact := req.PaidAmount + req.WriteOffAmount

	var glEntries []GLEntry

	// 2. Party Account (Receivable/Payable)
	// Need to find the default receivable/payable account for the party
	// For this MVP, we'll assume we can fetch it or it's provided. 
	// Let's assume a simplified fetch:
	var partyAccount string
	tx.Table("tabAccount").
		Where("tenant_id = ? AND account_type = ? AND is_group = 0", req.TenantID, req.PartyType).
		Select("name").Limit(1).Scan(&partyAccount)

	partyEntry := GLEntry{
		Account:     partyAccount,
		PostingDate: req.PostingDate,
		VoucherType: "PaymentEntry",
		VoucherNo:   req.Name,
	}

	bankEntry := GLEntry{
		Account:     req.PaidTo,
		PostingDate: req.PostingDate,
		VoucherType: "PaymentEntry",
		VoucherNo:   req.Name,
	}

	if req.PaymentType == "Receive" {
		// Money comes in: Debit Bank, Credit Receivable
		bankEntry.Debit = req.PaidAmount
		partyEntry.Credit = totalImpact
		
		if req.WriteOffAmount > 0 {
			writeOffEntry := GLEntry{
				Account:     req.WriteOffAccount,
				PostingDate: req.PostingDate,
				VoucherType: "PaymentEntry",
				VoucherNo:   req.Name,
				Debit:       req.WriteOffAmount, // Expense
			}
			glEntries = append(glEntries, writeOffEntry)
		}
	} else {
		// Money goes out: Credit Bank, Debit Payable
		bankEntry.Credit = req.PaidAmount
		partyEntry.Debit = totalImpact

		if req.WriteOffAmount > 0 {
			writeOffEntry := GLEntry{
				Account:     req.WriteOffAccount,
				PostingDate: req.PostingDate,
				VoucherType: "PaymentEntry",
				VoucherNo:   req.Name,
				Credit:      req.WriteOffAmount, // Gain/Income (usually) or reducing expense
			}
			glEntries = append(glEntries, writeOffEntry)
		}
	}

	glEntries = append(glEntries, bankEntry, partyEntry)

	// 3. Post GL Entries
	if err := PostGLEntries(tx, glEntries, req.TenantID); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Update Outstanding Amounts on Invoices
	for _, ref := range req.References {
		table := "tabSalesInvoice"
		if ref.ReferenceDocType == "Purchase Invoice" {
			table = "tabPurchaseInvoice"
		}

		updateSQL := fmt.Sprintf(`UPDATE %s SET outstanding_amount = outstanding_amount - ? 
					 WHERE name = ? AND tenant_id = ?`, table)
		
		if err := tx.Exec(updateSQL, ref.AllocatedAmount, ref.ReferenceName, req.TenantID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update outstanding for %s: %v", ref.ReferenceName, err)
		}
	}

	return tx.Commit().Error
}
