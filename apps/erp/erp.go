package erp

import (
	"github.com/goerp/goerp/apps/core/meta"
	"github.com/goerp/goerp/apps/erp/accounting"
	"github.com/goerp/goerp/apps/erp/buying"
	"github.com/goerp/goerp/apps/erp/hr"
	"github.com/goerp/goerp/apps/erp/manufacturing"
	"github.com/goerp/goerp/apps/erp/selling"
)

func InitERP() {
	// 1. Register Ledger Generators
	meta.DefaultLedgerEngine.RegisterGenerator("SalesInvoice", &selling.SalesInvoiceLedger{})

	// 2. Register all ERP hooks and events
	accounting.RegisterHooks()
	selling.RegisterHooks()
	buying.RegisterHooks()
	hr.RegisterHooks()
	manufacturing.RegisterHooks()
}
