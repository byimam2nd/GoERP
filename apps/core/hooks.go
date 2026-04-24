package core

import (
	"context"
	"github.com/goerp/goerp/apps/core/integration"
	"github.com/goerp/goerp/apps/core/meta"
	"github.com/goerp/goerp/apps/core/meta/migrator"
	"github.com/goerp/goerp/apps/core/meta/registry"
	"github.com/goerp/goerp/apps/core/meta/types"
	"github.com/goerp/goerp/apps/core/notification"
	"github.com/goerp/goerp/apps/core/script"
)

func RegisterCoreHooks() {
	// 1. Global Event Handlers (Webhooks & Server Scripts)
	events := []types.HookType{types.AfterInsert, types.AfterSave, types.OnSubmit, types.OnCancel}
	for _, event := range events {
		registry.DefaultHookRegistry.Register("*", event, GlobalCoreTrigger)
	}

	// 2. Unified Ledger Engine Hooks (Accounting & Stock)
	registry.DefaultHookRegistry.Register("*", types.OnSubmit, func(ctx *types.HookContext) error {
		return meta.DefaultLedgerEngine.Process(context.Background(), ctx.DocType.Name, ctx.Doc, false)
	})
	registry.DefaultHookRegistry.Register("*", types.OnCancel, func(ctx *types.HookContext) error {
		return meta.DefaultLedgerEngine.Process(context.Background(), ctx.DocType.Name, ctx.Doc, true)
	})

	// 3. Custom Field Auto-Migration
	registry.DefaultHookRegistry.Register("CustomField", types.AfterInsert, OnCustomFieldChange)
	registry.DefaultHookRegistry.Register("CustomField", types.AfterSave, OnCustomFieldChange)
}

func OnCustomFieldChange(ctx *types.HookContext) error {
	targetDocType := ctx.Doc["dt"].(string)
	tenantID, _ := ctx.Doc["tenant_id"].(string)

	dt, exists := registry.DefaultRegistry.Get(targetDocType, tenantID)
	if !exists {
		return nil
	}

	return migrator.MigrateDocType(dt)
}

func GlobalCoreTrigger(ctx *types.HookContext) error {
	tenantID, _ := ctx.Doc["tenant_id"].(string)

	// A. Run Server Scripts (Lua)
	script.RunServerScript(ctx.DocType.Name, string(ctx.HookType), ctx, tenantID)

	// B. Trigger Automated Notifications (Email/In-app)
	notification.TriggerNotifications(ctx.DocType.Name, string(ctx.HookType), tenantID, ctx.Doc)

	// C. Dispatch Webhooks
	integration.DispatchWebhooks(ctx.DocType.Name, string(ctx.HookType), ctx.Doc, tenantID)
	
	return nil
}
