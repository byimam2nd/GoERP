package validator

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"gorm.io/gorm"
	"math"
)

type InvariantFunc func(tx *gorm.DB, doctype, name, tenantID string) error

var registry = make(map[string][]InvariantFunc)

// RegisterInvariant adds a truth check for a specific DocType or "*" for all
func RegisterInvariant(doctype string, fn InvariantFunc) {
	registry[doctype] = append(registry[doctype], fn)
}

// ValidateInvariants runs all registered checks. If any fails, the transaction must rollback.
func ValidateInvariants(tx *gorm.DB, doctype, name, tenantID string) error {
	// 1. Run global invariants
	for _, fn := range registry["*"] {
		if err := fn(tx, doctype, name, tenantID); err != nil {
			return err
		}
	}

	// 2. Run DocType-specific invariants
	for _, fn := range registry[doctype] {
		if err := fn(tx, doctype, name, tenantID); err != nil {
			return err
		}
	}

	return nil
}

// --- CORE INVARIANTS ---

func InitCoreInvariants() {
	// 1. Accounting: Balanced Transaction
	RegisterInvariant("*", func(tx *gorm.DB, doctype, name, tenantID string) error {
		var result struct {
			Balance float64
		}
		// Check if this document has GL Entries and if they balance
		tx.Table("tabGLEntry").
			Select("SUM(debit) - SUM(credit) as balance").
			Where("voucher_type = ? AND voucher_no = ? AND tenant_id = ? AND docstatus = 1", doctype, name, tenantID).
			Scan(&result)

		if math.Abs(result.Balance) > 0.0001 {
			return fmt.Errorf("Accounting Invariant Violated: Transaction is not balanced (Difference: %v)", result.Balance)
		}
		return nil
	})

	// 2. Stock: No Negative Stock (Optional per Tenant/Item)
	// This would check 'tabBin' after the transaction
}

// Precision Rounding to avoid float issues
func Round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
