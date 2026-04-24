package accounting

import (
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"github.com/goerp/goerp/apps/erp/accounting/service"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("SalesInvoice", types.OnSubmit, HandleSalesInvoiceSubmit)
	registry.DefaultHookRegistry.Register("PaymentEntry", types.OnSubmit, OnPaymentEntrySubmit)
	registry.DefaultHookRegistry.Register("JournalEntry", types.OnSubmit, HandleJournalEntrySubmit)
}

func OnPaymentEntrySubmit(ctx *types.HookContext) error {
	tenantID := ctx.Doc["tenant_id"].(string)
	party := ctx.Doc["party"].(string)
	amount, _ := ctx.Doc["paid_amount"].(float64)
	bankAccount := ctx.Doc["paid_to"].(string)

	// 1. Auto-Reconciliation (FIFO matching with Outstanding Invoices)
	remainingAmount := amount
	var invoices []map[string]interface{}
	database.DB.Table("tabSalesInvoice").
		Where("customer = ? AND outstanding_amount > 0 AND docstatus = 1 AND tenant_id = ?", party, tenantID).
		Order("posting_date ASC").
		Find(&invoices)

	for _, inv := range invoices {
		if remainingAmount <= 0 { break }

		outstanding := inv["outstanding_amount"].(float64)
		paymentForThisInvoice := 0.0

		if remainingAmount >= outstanding {
			paymentForThisInvoice = outstanding
			remainingAmount -= outstanding
		} else {
			paymentForThisInvoice = remainingAmount
			remainingAmount = 0
		}

		database.DB.Table("tabSalesInvoice").Where("name = ?", inv["name"]).
			Update("outstanding_amount", outstanding - paymentForThisInvoice)
		
		meta.CreateDocLink("SalesInvoice", inv["name"].(string), "PaymentEntry", ctx.Doc["name"].(string), tenantID)
	}

	// 2. Post GL Entries
	glEntries := []service.GLEntry{
		{Account: bankAccount, Debit: amount, Credit: 0, VoucherType: "PaymentEntry", VoucherNo: ctx.Doc["name"].(string)},
		{Account: "Accounts Receivable", Debit: 0, Credit: amount, VoucherType: "PaymentEntry", VoucherNo: ctx.Doc["name"].(string)},
	}

	return service.PostGLEntries(glEntries, tenantID)
}

func HandleSalesInvoiceSubmit(ctx *types.HookContext) error {
	tenantID := ctx.Doc["tenant_id"].(string)
	invoiceName := ctx.Doc["name"].(string)
	customerName := ctx.Doc["customer"].(string)
	total, _ := ctx.Doc["total"].(float64)

	// Set Outstanding Amount = Total at first
	database.DB.Table("tabSalesInvoice").Where("name = ?", invoiceName).Update("outstanding_amount", total)

	glEntries := []service.GLEntry{
		{Account: "Accounts Receivable", Debit: total, Credit: 0, Against: customerName, VoucherType: "SalesInvoice", VoucherNo: invoiceName},
		{Account: "Sales Income", Debit: 0, Credit: total, Against: customerName, VoucherType: "SalesInvoice", VoucherNo: invoiceName},
	}

	return service.PostGLEntries(glEntries, tenantID)
}

func HandleJournalEntrySubmit(ctx *types.HookContext) error {
	tenantID := ctx.Doc["tenant_id"].(string)
	jeName := ctx.Doc["name"].(string)

	var entries []service.GLEntry
	
	// Fetch Journal Entry Accounts (Child Table)
	var accounts []map[string]interface{}
	database.DB.Table("tabJournalEntryAccount").
		Where("parent = ? AND tenant_id = ?", jeName, tenantID).
		Find(&accounts)

	for _, acc := range accounts {
		debit, _ := acc["debit"].(float64)
		credit, _ := acc["credit"].(float64)
		
		entries = append(entries, service.GLEntry{
			Account:     acc["account"].(string),
			Debit:       debit,
			Credit:      credit,
			VoucherType: "JournalEntry",
			VoucherNo:   jeName,
		})
	}

	return service.PostGLEntries(entries, tenantID)
}
