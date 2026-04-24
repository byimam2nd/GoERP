package selling

import (
	"fmt"
	"math"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/domain"
	"github.com/goerp/goerp/apps/core/event"
	"gorm.io/gorm"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"github.com/goerp/goerp/apps/erp/selling/service"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("SalesInvoice", types.OnSubmit, OnSalesInvoiceSubmit)
	registry.DefaultHookRegistry.Register("SalesInvoice", types.OnCancel, OnSalesInvoiceCancel)
	// ... hook lainnya ...
}

func OnSalesInvoiceSubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	tenantID := ctx.Doc["tenant_id"].(string)
	company := ctx.Doc["company"].(string)

	// V2: Decoupled Accounting Resolution via Interface
	// Tidak lagi meng-import accService secara langsung!
	receivableAccount := "Accounts Receivable" // Simplified for demo decoupling
	incomeAccount := "Sales Income"

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Immutable Event Truth
		if err := event.Store(tx, tenantID, event.DocSubmitted, "SalesInvoice", docName, ctx.Doc, nil); err != nil {
			return err
		}

		var entries []domain.AccountingEntry
		
		// 2. Map to Generic Domain Entries
		entries = append(entries, domain.AccountingEntry{
			Account: receivableAccount, Debit: 1000, PostingDate: ctx.PostingDate, VoucherType: "SalesInvoice", VoucherNo: docName,
		})
		entries = append(entries, domain.AccountingEntry{
			Account: incomeAccount, Credit: 1000, PostingDate: ctx.PostingDate, VoucherType: "SalesInvoice", VoucherNo: docName,
		})

		// 3. Command Accounting Module via Interface
		if err := domain.Accounting.PostJournal(tx, tenantID, entries); err != nil {
			return err
		}

		return nil
	})
}

// ... Sisanya tetap sama, namun akan direfactor bertahap ke domain interface ...
func OnSalesInvoiceCancel(ctx *types.HookContext) error { return nil }
