package script

import (
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/goerp/goerp/apps/core/meta/types"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func RunServerScript(doctype string, event string, ctx *types.HookContext, tenantID string) error {
	var scripts []map[string]interface{}
	err := database.DB.Table("tabServerScript").
		Where("ref_doctype = ? AND event = ? AND is_active = ? AND tenant_id = ?", doctype, event, true, tenantID).
		Find(&scripts).Error

	if err != nil || len(scripts) == 0 {
		return nil
	}

	L := lua.NewState()
	defer L.Close()

	// Create 'doc' table in Lua to allow access/modification
	docTable := L.NewTable()
	for k, v := range ctx.Doc {
		L.SetTable(docTable, lua.LString(k), convertToGoLuaValue(L, v))
	}
	L.SetGlobal("doc", docTable)

	for _, s := range scripts {
		code := s["script"].(string)
		if err := L.DoString(code); err != nil {
			logger.Log.Error("Lua Script Error", zap.String("script", s["script_name"].(string)), zap.Error(err))
			return err
		}
	}

	// Update back Go map from Lua table
	docTable.ForEach(func(k, v lua.LValue) {
		ctx.Doc[k.String()] = convertToGOValue(v)
	})

	return nil
}

func convertToGoLuaValue(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case string:
		return lua.LString(val)
	case float64:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	default:
		return lua.LNil
	}
}

func convertToGOValue(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTString:
		return v.String()
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTBool:
		return bool(v.(lua.LBool))
	default:
		return nil
	}
}
