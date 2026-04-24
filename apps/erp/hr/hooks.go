package hr

import (
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	accService "github.com/goerp/goerp/apps/erp/accounting/service"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("PayrollEntry", types.OnSubmit, OnPayrollEntrySubmit)
}

func OnPayrollEntrySubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	tenantID := ctx.Doc["tenant_id"].(string)
	company := ctx.Doc["company"].(string)

	// Fetch all Salary Slips for this Payroll Entry
	var slips []map[string]interface{}
	if err := database.DB.Table("tabSalarySlip").
		Where("payroll_entry = ? AND tenant_id = ?", docName, tenantID).
		Find(&slips).Error; err != nil {
		return err
	}

	var glEntries []accService.GLEntry

	for _, slip := range slips {
		employee := slip["employee"].(string)
		netPay, _ := slip["net_pay"].(float64)

		// Get Salary Structure to find accounts
		var structData map[string]interface{}
		err := database.DB.Table("tabSalaryStructure").
			Where("employee = ? AND company = ? AND tenant_id = ? AND is_active = ?", employee, company, tenantID, true).
			First(&structData).Error
		
		if err != nil {
			continue // Skip if no active structure found
		}

		expenseAccount := structData["salary_expense_account"].(string)
		payableAccount := structData["payroll_payable_account"].(string)

		// Debit Expense
		glEntries = append(glEntries, accService.GLEntry{
			Account:     expenseAccount,
			Debit:       netPay,
			Credit:      0,
			VoucherType: "PayrollEntry",
			VoucherNo:   docName,
		})

		// Credit Payable
		glEntries = append(glEntries, accService.GLEntry{
			Account:     payableAccount,
			Debit:       0,
			Credit:      netPay,
			VoucherType: "PayrollEntry",
			VoucherNo:   docName,
		})
	}

	if len(glEntries) > 0 {
		return accService.PostGLEntries(glEntries, tenantID)
	}

	return nil
}
