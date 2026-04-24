package stock

import (
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"github.com/goerp/goerp/apps/erp/stock/service"
	"github.com/goerp/goerp/apps/core/database"
)

func init() {
	registry.DefaultHookRegistry.Register("StockEntry", types.OnSubmit, HandleStockEntrySubmit)
}

func HandleStockEntrySubmit(ctx *types.HookContext) error {
	tenantID := ctx.Doc["tenant_id"].(string)
	docName := ctx.Doc["name"].(string)

	var lines []map[string]interface{}
	database.DB.Table("tabStockEntryDetail").
		Where("parent = ? AND tenant_id = ?", docName, tenantID).
		Find(&lines)

	for _, line := range lines {
		itemCode := line["item_code"].(string)
		qty, _ := line["qty"].(float64)
		sWarehouse, _ := line["s_warehouse"].(string)
		tWarehouse, _ := line["t_warehouse"].(string)
		serialNo, _ := line["serial_no"].(string)
		batchNo, _ := line["batch_no"].(string)

		// Outgoing from source
		if sWarehouse != "" {
			service.UpdateStock(service.StockInput{
				ItemCode: itemCode, Warehouse: sWarehouse, Qty: -qty, 
				VoucherType: "StockEntry", VoucherNo: docName, 
				SerialNo: serialNo, BatchNo: batchNo, TenantID: tenantID,
			})
		}

		// Incoming to target
		if tWarehouse != "" {
			service.UpdateStock(service.StockInput{
				ItemCode: itemCode, Warehouse: tWarehouse, Qty: qty, 
				VoucherType: "StockEntry", VoucherNo: docName, 
				SerialNo: serialNo, BatchNo: batchNo, TenantID: tenantID,
			})
		}
	}
	return nil
}
