package service

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type AccountStatement struct {
	AccountName    string              `json:"account_name"`
	ParentAccount  string              `json:"parent_account"`
	IsGroup        bool                `json:"is_group"`
	RootType       string              `json:"root_type"`
	OpeningBalance float64             `json:"opening_balance"`
	Debit          float64             `json:"debit"`
	Credit         float64             `json:"credit"`
	ClosingBalance float64             `json:"closing_balance"`
	Children       []*AccountStatement `json:"children,omitempty"`
}

// GetFinancialStatement is a generic engine for Trial Balance, P&L, or Balance Sheet
func GetFinancialStatement(tenantID, company, reportType string, startDate, endDate time.Time) ([]*AccountStatement, error) {
	// 1. Fetch Accounts based on reportType
	var accounts []map[string]interface{}
	db := database.DB.Table("tabAccount").Where("tenant_id = ? AND company = ?", tenantID, company)
	if reportType != "Trial Balance" {
		db = db.Where("report_type = ?", reportType)
	}
	if err := db.Find(&accounts).Error; err != nil {
		return nil, err
	}

	// 2. Fetch Aggregated Period Data
	type PeriodData struct {
		Account string
		Debit   float64
		Credit  float64
	}
	var periodData []PeriodData
	database.DB.Table("tabGLEntry").
		Select("account, SUM(debit) as debit, SUM(credit) as credit").
		Where("tenant_id = ? AND company = ? AND posting_date BETWEEN ? AND ? AND docstatus = 1", tenantID, company, startDate, endDate).
		Group("account").
		Scan(&periodData)

	// 3. Fetch Opening Balance
	type OpeningData struct {
		Account string
		Opening float64
	}
	var openingData []OpeningData
	database.DB.Table("tabGLEntry").
		Select("account, SUM(debit - credit) as opening").
		Where("tenant_id = ? AND company = ? AND posting_date < ? AND docstatus = 1", tenantID, company, startDate).
		Group("account").
		Scan(&openingData)

	periodMap := make(map[string]PeriodData)
	for _, d := range periodData { periodMap[d.Account] = d }
	openingMap := make(map[string]float64)
	for _, d := range openingData { openingMap[d.Account] = d.Opening }

	nodes := make(map[string]*AccountStatement)
	for _, acc := range accounts {
		name := acc["name"].(string)
		pm := periodMap[name]
		opening := openingMap[name]
		
		nodes[name] = &AccountStatement{
			AccountName:    name,
			ParentAccount:  acc["parent_account"].(string),
			IsGroup:        acc["is_group"].(bool),
			RootType:       acc["root_type"].(string),
			OpeningBalance: opening,
			Debit:          pm.Debit,
			Credit:         pm.Credit,
			ClosingBalance: opening + pm.Debit - pm.Credit,
		}
	}

	// Adjacency list for hierarchy
	childrenMap := make(map[string][]string)
	for name, node := range nodes {
		if node.ParentAccount != "" {
			childrenMap[node.ParentAccount] = append(childrenMap[node.ParentAccount], name)
		}
	}

	// Recursive Rollup
	var rollup func(string) *AccountStatement
	rollup = func(name string) *AccountStatement {
		node := nodes[name]
		if !node.IsGroup {
			return node
		}

		node.Children = []*AccountStatement{}
		var totalOpening, totalDebit, totalCredit, totalClosing float64
		for _, childName := range childrenMap[name] {
			childNode := rollup(childName)
			totalOpening += childNode.OpeningBalance
			totalDebit += childNode.Debit
			totalCredit += childNode.Credit
			totalClosing += childNode.ClosingBalance
			node.Children = append(node.Children, childNode)
		}
		
		node.OpeningBalance = totalOpening
		node.Debit = totalDebit
		node.Credit = totalCredit
		node.ClosingBalance = totalClosing
		return node
	}

	var rootNodes []*AccountStatement
	for name, node := range nodes {
		// Roots are nodes whose parents are NOT in our current filtered 'nodes' map
		if node.ParentAccount == "" || nodes[node.ParentAccount] == nil {
			rootNodes = append(rootNodes, rollup(name))
		}
	}

	return rootNodes, nil
}

func GetTrialBalance(tenantID, company string, start, end time.Time) ([]*AccountStatement, error) {
	return GetFinancialStatement(tenantID, company, "Trial Balance", start, end)
}

func GetProfitAndLoss(tenantID, company string, start, end time.Time) ([]*AccountStatement, error) {
	return GetFinancialStatement(tenantID, company, "Profit and Loss", start, end)
}

func GetBalanceSheet(tenantID, company string, start, end time.Time) ([]*AccountStatement, error) {
	// Note: For a real Balance Sheet, we might need to inject "Provisional Profit/Loss" 
	// from P&L into Equity if the period is not yet closed.
	return GetFinancialStatement(tenantID, company, "Balance Sheet", start, end)
}
