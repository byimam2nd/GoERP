package meta

import "context"

// Entry represents a generic ledger entry (GL or Stock)
type LedgerEntry interface {
	GetVoucherInfo() (string, string) // Returns VoucherType, VoucherNo
}

// GLEntry represents a specific General Ledger record
type GLEntry struct {
	Account   string
	Debit     float64
	Credit    float64
	Against   string
	Remarks   string
}

// StockEntry represents a specific Stock Ledger record
type StockEntry struct {
	ItemCode  string
	Warehouse string
	Qty       float64
	Rate      float64
}

// LedgerGenerator is the interface that DocTypes must implement 
// to support automated accounting/stock effects.
type LedgerGenerator interface {
	GenerateGL(ctx context.Context, doc map[string]interface{}) ([]GLEntry, error)
	GenerateStock(ctx context.Context, doc map[string]interface{}) ([]StockEntry, error)
}
