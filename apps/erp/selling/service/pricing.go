package service

import (
	"github.com/goerp/goerp/apps/core/database"
	"time"
)

type PricingContext struct {
	ItemCode      string
	ItemGroup     string
	Customer      string
	CustomerGroup string
	Qty           float64
	PriceList     string
	Currency      string
	TransactionDate time.Time
	TenantID      string
}

type PricingResult struct {
	BaseRate          float64
	PriceListRate     float64
	DiscountPercentage float64
	DiscountAmount     float64
	FinalRate         float64
	RuleName          string
}

func ResolvePrice(ctx PricingContext) (PricingResult, error) {
	if ctx.TransactionDate.IsZero() {
		ctx.TransactionDate = time.Now()
	}

	var rules []map[string]interface{}
	result := PricingResult{}

	// 1. Fetch Applicable Rules
	// Logic: Filter by date, tenant, and common filters. More specific logic is handled in code.
	query := database.DB.Table("tabPricingRule").
		Where("is_active = ? AND tenant_id = ?", true, ctx.TenantID).
		Where("valid_from <= ? AND (valid_upto >= ? OR valid_upto IS NULL)", ctx.TransactionDate, ctx.TransactionDate).
		Order("priority DESC, creation DESC")

	query.Find(&rules)

	// 2. Evaluate each rule against the context
	for _, rule := range rules {
		if !isRuleApplicable(rule, ctx) {
			continue
		}

		result.RuleName = rule["name"].(string)

		// Handle Fixed Price (Overwrites everything)
		if fixed, ok := rule["fixed_price"].(float64); ok && fixed > 0 {
			result.FinalRate = fixed
			return result, nil
		}

		// Handle Discount Percentage
		if discP, ok := rule["discount_percentage"].(float64); ok && discP > 0 {
			result.DiscountPercentage = discP
		}

		// Handle Discount Amount
		if discA, ok := rule["discount_amount"].(float64); ok && discA > 0 {
			result.DiscountAmount = discA
		}

		// ERPNext style: Usually the first high-priority rule wins. 
		// If stacking is needed, we wouldn't return here.
		break 
	}

	// 3. Final Calculation logic (Placeholder for Price List integration)
	// In a full system, we would fetch Item Price first.
	return result, nil
}

func isRuleApplicable(rule map[string]interface{}, ctx PricingContext) bool {
	// Check Item
	if rule["item_code"] != nil && rule["item_code"] != "" && rule["item_code"] != ctx.ItemCode {
		return false
	}
	if rule["item_group"] != nil && rule["item_group"] != "" && rule["item_group"] != ctx.ItemGroup {
		return false
	}

	// Check Customer
	if rule["customer"] != nil && rule["customer"] != "" && rule["customer"] != ctx.Customer {
		return false
	}
	if rule["customer_group"] != nil && rule["customer_group"] != "" && rule["customer_group"] != ctx.CustomerGroup {
		return false
	}

	// Check Quantity Tier
	if minQty, ok := rule["min_qty"].(float64); ok && ctx.Qty < minQty {
		return false
	}
	if maxQty, ok := rule["max_qty"].(float64); ok && maxQty > 0 && ctx.Qty > maxQty {
		return false
	}

	return true
}
