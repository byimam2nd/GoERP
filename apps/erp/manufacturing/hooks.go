package manufacturing

import (
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"go.uber.org/zap"
)

func RegisterHooks() {
	registry.DefaultHookRegistry.Register("WorkOrder", types.OnSubmit, OnWorkOrderSubmit)
}

func OnWorkOrderSubmit(ctx *types.HookContext) error {
	docName := ctx.Doc["name"].(string)
	item := ctx.Doc["production_item"].(string)
	qty, _ := ctx.Doc["qty"].(float64)

	logger.Log.Info("Work Order submitted", 
		zap.String("name", docName), 
		zap.String("item", item), 
		zap.Float64("qty", qty))
	
	// Future: Trigger automatic Stock Entry for material issue
	return nil
}
