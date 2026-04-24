package selling

import (
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	accService "github.com/goerp/goerp/apps/erp/accounting/service"
	"github.com/goerp/goerp/apps/erp/selling/service"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("SalesInvoice", types.OnSubmit, OnSalesInvoiceSubmit)
	registry.DefaultHookRegistry.Register("SalesInvoice", types.BeforeInsert, ApplyPricingOnInsert)
}

func ApplyPricingOnInsert(ctx *types.HookContext) error {
	tenantID := ctx.Doc["tenant_id"].(string)
	customerGroup := "Individual" // Default, ideally fetch from Customer Link

	// Process each item in the child table (simplified)
	if items, ok := ctx.Doc["items"].([]map[string]interface{}); ok {
		for _, item := range items {
			discount, _ := service.EvaluatePricingRule(service.PricingContext{
				ItemCode:      item["item_code"].(string),
				CustomerGroup: customerGroup,
				Qty:           item["qty"].(float64),
				TenantID:      tenantID,
			})

			if discount > 0 {
				item["discount_percentage"] = discount
				// Recalculate amount logic would go here
			}
		}
	}
	return nil
	}


func OnSalesInvoiceSubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	tenantID := ctx.Doc["tenant_id"].(string)
	
	grandTotal, _ := ctx.Doc["grand_total"].(float64)
	netTotal, _ := ctx.Doc["net_total"].(float64)
	receivableAccount := ctx.Doc["debit_to"].(string)
	incomeAccount := ctx.Doc["income_account"].(string)

	var glEntries []accService.GLEntry

	// 1. Debit Receivable (Full Amount)
	glEntries = append(glEntries, accService.GLEntry{
		Account:     receivableAccount,
		Debit:       grandTotal,
		Credit:      0,
		VoucherType: "SalesInvoice",
		VoucherNo:   docName,
	})

	// 2. Credit Income (Net Amount)
	glEntries = append(glEntries, accService.GLEntry{
		Account:     incomeAccount,
		Debit:       0,
		Credit:      netTotal,
		VoucherType: "SalesInvoice",
		VoucherNo:   docName,
	})

	// 3. Credit Each Tax Component
	if taxes, ok := ctx.Doc["taxes"].([]interface{}); ok {
		for _, tRaw := range taxes {
			t := tRaw.(map[string]interface{})
			taxAccount := t["account_head"].(string)
			taxAmount := t["tax_amount"].(float64)

			if taxAmount > 0 {
				glEntries = append(glEntries, accService.GLEntry{
					Account:     taxAccount,
					Debit:       0,
					Credit:      taxAmount,
					VoucherType: "SalesInvoice",
					VoucherNo:   docName,
				})
			}
		}
	}

	return accService.PostGLEntries(glEntries, tenantID)
}
