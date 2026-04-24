package buying

import (
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	accService "github.com/goerp/goerp/apps/erp/accounting/service"
	stockService "github.com/goerp/goerp/apps/erp/stock/service"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("PurchaseReceipt", types.OnSubmit, OnPurchaseReceiptSubmit)
	registry.DefaultHookRegistry.Register("PurchaseInvoice", types.OnSubmit, OnPurchaseInvoiceSubmit)
}

func OnPurchaseReceiptSubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	tenantID := ctx.Doc["tenant_id"].(string)

	// Fetch detail lines
	var lines []map[string]interface{}
	if err := database.DB.Table("tabPurchaseReceiptItem").
		Where("parent = ? AND tenant_id = ?", docName, tenantID).
		Find(&lines).Error; err != nil {
		return err
	}

	for _, line := range lines {
		itemCode := line["item_code"].(string)
		qty := line["qty"].(float64)
		rate := line["rate"].(float64)
		warehouse := line["warehouse"].(string)

		// Incoming stock
		err := stockService.UpdateStock(stockService.StockInput{
			ItemCode:    itemCode,
			Warehouse:   warehouse,
			Qty:         qty,
			Rate:        rate,
			VoucherType: "PurchaseReceipt",
			VoucherNo:   docName,
			TenantID:    tenantID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func OnPurchaseInvoiceSubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	tenantID := ctx.Doc["tenant_id"].(string)
	total, _ := ctx.Doc["total"].(float64)
	payableAccount := ctx.Doc["credit_to"].(string)
	expenseAccount := ctx.Doc["expense_account"].(string)

	var glEntries []accService.GLEntry

	// Debit Expense/Asset
	glEntries = append(glEntries, accService.GLEntry{
		Account:     expenseAccount,
		Debit:       total,
		Credit:      0,
		VoucherType: "PurchaseInvoice",
		VoucherNo:   docName,
	})

	// Credit Payable
	glEntries = append(glEntries, accService.GLEntry{
		Account:     payableAccount,
		Debit:       0,
		Credit:      total,
		VoucherType: "PurchaseInvoice",
		VoucherNo:   docName,
	})

	return accService.PostGLEntries(glEntries, tenantID)
}
