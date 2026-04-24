package types

type HookType string

const (
	BeforeInsert HookType = "BeforeInsert"
	AfterInsert  HookType = "AfterInsert"
	BeforeSave   HookType = "BeforeSave"
	AfterSave    HookType = "AfterSave"
	BeforeDelete HookType = "BeforeDelete"
	AfterDelete  HookType = "AfterDelete"
	OnSubmit     HookType = "OnSubmit"
	OnCancel     HookType = "OnCancel"
)

type HookContext struct {
	DocType  *DocType
	Doc      map[string]interface{}
	User     string
	HookType HookType
}

type HookFunc func(ctx *HookContext) error
