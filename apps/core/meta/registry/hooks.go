package registry

import (
	"github.com/goerp/goerp/apps/core/meta/types"
	"github.com/goerp/goerp/apps/core/integration"
	"sync"
)

type HookRegistry struct {
	hooks map[string]map[types.HookType][]types.HookFunc
	mu    sync.RWMutex
}

var DefaultHookRegistry = &HookRegistry{
	hooks: make(map[string]map[types.HookType][]types.HookFunc),
}

func (r *HookRegistry) Register(doctype string, hookType types.HookType, fn types.HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hooks[doctype] == nil {
		r.hooks[doctype] = make(map[types.HookType][]types.HookFunc)
	}
	r.hooks[doctype][hookType] = append(r.hooks[doctype][hookType], fn)
}

func (r *HookRegistry) Trigger(doctype string, hookType types.HookType, ctx *types.HookContext) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Run specific hooks
	if hooks, exists := r.hooks[doctype]; exists {
		for _, fn := range hooks[hookType] {
			if err := fn(ctx); err != nil {
				return err
			}
		}
	}

	// 2. Trigger External Webhooks
	tenantID, _ := ctx.Doc["tenant_id"].(string)
	integration.TriggerWebhook(doctype, string(hookType), ctx.Doc, tenantID)

	// 3. Run wildcard hooks
	if hooks, exists := r.hooks["*"]; exists {
		for _, fn := range hooks[hookType] {
			if err := fn(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
